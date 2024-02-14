package rinha

import (
	"context"
	"fmt"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/influxdb2"
	"rinha-be-2-app/internal/metric"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const (
	opCredito string = "c"
	opDebito  string = "d"
)

// TransacaoEventoT representa um evento de transação a ser lançado no InfluxDB2
type TransacaoEventoT struct {
	Timestamp time.Time     // Timestamp do evento
	Tipo      string        // Tipo de transação (ex: "c", "d")
	Cliente   string        // Código do cliente
	Descricao string        // Descrição da transação
	Duracao   time.Duration // Duração do processamento da transação
	Valor     int64         // Valor da transação
	Metrica   string        // Nome da métrica
}

// TransacaoT representa o payload da requisição de transação
type TransacaoT struct {
	Cliente   string    `json:"-"`
	Valor     int64     `json:"valor"`
	Tipo      string    `json:"tipo"`
	Descricao string    `json:"descricao"`
	Timestamp time.Time `json:"realizada_em"`
}

type TransacaoErroI interface {
	Error()
	As()
}

type TransacaoErroLimiteUltrapassadoT struct {
	Erro   string
	Codigo int
}

func (e *TransacaoErroLimiteUltrapassadoT) Error() string {
	return fmt.Sprintf("Erro: %s, Código: %d", e.Erro, e.Codigo)
}

type TransacaoErroTipoInvalidoT struct {
	Erro   string
	Codigo int
}

func (e *TransacaoErroTipoInvalidoT) Error() string {
	return fmt.Sprintf("Erro: %s, Código: %d", e.Erro, e.Codigo)
}

func NewTransacaoErroLimiteUltrapassado() *TransacaoErroLimiteUltrapassadoT {
	return &TransacaoErroLimiteUltrapassadoT{
		Erro:   "Limite de saldo ultrapassado",
		Codigo: 100,
	}
}

func NewTransacaoErroTipoInvalido() *TransacaoErroTipoInvalidoT {
	return &TransacaoErroTipoInvalidoT{
		Erro:   "Tipo de transação inválido",
		Codigo: 101,
	}
}

// Extrato representa o payload da resposta da requisição de extrato
type ExtratoT struct {
	Saldo struct {
		Total       int64  `json:"total"`
		DataExtrato string `json:"data_extrato"`
		Limite      int64  `json:"limite"`
	} `json:"saldo"`
	UltimasTransacoes []TransacaoT `json:"ultimas_transacoes"`
}

func (e TransacaoEventoT) GetDataPoint() (string, metric.TagT, metric.FieldT, time.Time) {
	t := make(metric.TagT, 0)
	t["tipo"] = e.Tipo
	t["cliente"] = e.Cliente

	f := make(metric.FieldT, 0)
	f["valor"] = e.Valor
	f["duracao"] = e.Duracao.Nanoseconds()
	f["descricao"] = e.Descricao

	return e.Metrica, t, f, e.Timestamp
}

// ================================================================================================================

func ExecutaTransacao(ctx context.Context, t TransacaoT) (*ClienteSaldoT, error) {

	evento := TransacaoEventoT{
		Timestamp: time.Now(),
		Tipo:      t.Tipo,
		Cliente:   ctx.Value("clienteID").(string),
		Descricao: t.Descricao,
		Valor:     t.Valor,
		Metrica:   "transacoes",
	}

	clienteSaldo, err := ConsultaSaldo(ctx)
	if err != nil {
		return nil, err
	}

	switch t.Tipo {
	case opCredito:
		clienteSaldo.Saldo += t.Valor
	case opDebito:
		saldoAtual := clienteSaldo.Saldo
		saldoAtual -= t.Valor
		if saldoAtual < clienteSaldo.Cliente.Limite*-1 {
			return nil, NewTransacaoErroLimiteUltrapassado()
		}
		clienteSaldo.Saldo = saldoAtual
	default:
		return nil, NewTransacaoErroTipoInvalido()
	}

	evento.Duracao = time.Since(evento.Timestamp)

	err = influxdb2.Write(config.AppConfig.InfluxDB2Org, config.AppConfig.InfluxDB2Bucket, evento)
	if err != nil {
		return nil, err
	}

	spew.Dump(evento)

	return clienteSaldo, nil

}
