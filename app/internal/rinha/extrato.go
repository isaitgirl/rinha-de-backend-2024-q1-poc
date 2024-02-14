package rinha

import (
	"context"
	"rinha-be-2-app/internal/influxdb2"
	"rinha-be-2-app/internal/logger"
	"time"
)

func MontaExtrato(ctx context.Context) (ExtratoT, error) {

	clienteID := ctx.Value("clienteID").(string)
	logger.Info("Montando extrato do cliente", "cliente", clienteID, "uuid", ctx.Value("uuid").(string))

	// Busca o saldo
	clienteSaldo, err := ConsultaSaldo(ctx)
	if err != nil {
		logger.Warn("Erro ao buscar saldo", "cliente", clienteID, "erro", err, "uuid", ctx.Value("uuid").(string))
		return ExtratoT{}, err
	}

	extrato := ExtratoT{}
	extrato.Saldo.Total = clienteSaldo.Saldo
	extrato.Saldo.Limite = clienteSaldo.Cliente.Limite
	extrato.Saldo.DataExtrato = time.Now().Format(time.RFC3339Nano)

	// As últimas transações
	transacoes, err := ConsultaUltimasTransacoes(ctx)
	if err != nil {
		logger.Warn("Erro ao buscar transações", "cliente", clienteID, "erro", err, "uuid", ctx.Value("uuid").(string))
		return ExtratoT{}, err
	}

	extrato.UltimasTransacoes = transacoes

	return extrato, nil
}

func ConsultaUltimasTransacoes(ctx context.Context) ([]TransacaoT, error) {

	clienteID := ctx.Value("clienteID").(string)
	uuid := ctx.Value("uuid").(string)

	txStart := time.Now()

	logger.Info("Consultando últimas transações", "cliente", clienteID, "uuid", uuid)
	defer func() {
		logger.Info("Consulta de transações finalizada", "cliente", clienteID, "duracao", time.Since(txStart), "uuid", uuid)
	}()

	query := `from(bucket: "rinha")
						|> range(start: -7d, stop: now())
						|> filter(fn: (r) => r["_measurement"] == "transacoes")
						|> filter(fn: (r) => r["cliente"] == "` + clienteID + `")
						|> pivot(rowKey:["_time","cliente"], columnKey: ["_field"], valueColumn: "_value")
						|> filter(fn: (r) => r["descricao"] != "Abertura de conta")
						|> group(columns: ["cliente"])
						|> drop(columns: ["_start","_stop"])
						|> sort(columns: ["_time"], desc: true)
						|> limit(n: 10)						
						|> yield(name: "transacoes")
		`

	result, err := influxdb2.Query(query)
	if err != nil {
		logger.Warn("Erro ao consultar transações", "erro", err)
		return nil, err
	}

	transacoes := make([]TransacaoT, 0)

	for result.Next() {
		transacoes = append(transacoes, TransacaoT{
			Tipo:      result.Record().ValueByKey("tipo").(string),
			Valor:     result.Record().ValueByKey("valor").(int64),
			Descricao: result.Record().ValueByKey("descricao").(string),
			Timestamp: result.Record().Time(),
		})
	}
	// check for an error
	if result.Err() != nil {
		logger.Warn("Erro ao iterar o resultado", "erro", result.Err())
		return nil, result.Err()
	}

	return transacoes, nil
}

// ConsultaSaldo retorna o saldo, o limite e um erro (se houver)
func ConsultaSaldo(ctx context.Context) (*ClienteSaldoT, error) {

	clienteID := ctx.Value("clienteID").(string)
	uuid := ctx.Value("uuid").(string)

	logger.Info("Consultando saldo do cliente", "cliente", clienteID, "uuid", uuid)

	//TODO: neste ponto, implementar o mutex do etcd para garantir que o saldo seja verificado corretamente

	txStart := time.Now()
	defer func() {
		logger.Info("Consulta de saldo finalizada", "cliente", clienteID, "duracao", time.Since(txStart), "uuid", uuid)
	}()

	// Verifica o saldo atual

	query := `import "join"
						transacoes = from(bucket: "rinha")
						|> range(start: -7d, stop: now())
						|> filter(fn: (r) => r["_measurement"] == "transacoes")
						|> filter(fn: (r) => r["cliente"] == "` + clienteID + `")
						|> pivot(rowKey:["_time","cliente"], columnKey: ["_field"], valueColumn: "_value")
						|> group(columns: ["cliente","tipo"])
						|> drop(columns: ["_start","_stop"])
						
						creditos = transacoes 
						|> filter(fn: (r) => r["tipo"] == "c")
						|> sum(column: "valor")
						|> drop(columns: ["tipo"])
						
						debitos = transacoes 
						|> filter(fn: (r) => r["tipo"] == "d")
						|> sum(column: "valor")
						|> drop(columns: ["tipo"])
						
						cliente = from(bucket: "rinha")
						|> range(start: -7d, stop: now())
						|> filter(fn: (r) => r["_measurement"] == "clientes")
						|> filter(fn: (r) => r["cliente"] == "` + clienteID + `")
						|> filter(fn: (r) => r["_field"] == "limite")
						|> keep(columns: ["cliente","_value"])
						
						saldo = join.inner(
								left: creditos,
								right: debitos,
								on: (l,r) => l.cliente == r.cliente,
								as: (l,r) => ({
										cliente: l.cliente,
										total: l.valor - r.valor,
								})
						)
						|> join.inner(
							right: cliente,
							on: (l,r) => l.cliente == r.cliente,
							as: (l,r) => ({
								cliente: l.cliente,
								total: l.total,
								limite: r._value,
							})
						)
						|> yield(name: "saldo")
						`

	result, err := influxdb2.Query(query)
	if err != nil {
		logger.Warn("Erro ao consultar saldo", "cliente", clienteID, "erro", err, "uuid", uuid)
		return nil, err
	}

	saldoAtual := int64(0)
	limite := int64(0)

	// Saldo possui apenas um registro
	for result.Next() {
		saldoAtual = result.Record().ValueByKey("total").(int64)
		limite = result.Record().ValueByKey("limite").(int64)
		logger.Info("Saldo consultado", "cliente", clienteID, "saldoAtual", saldoAtual, "limite", limite, "uuid", uuid)
	}
	// check for an error
	if result.Err() != nil {
		logger.Warn("Erro ao iterar o resultado da consulta de saldo", "cliente", clienteID, "erro", result.Err(), "uuid", uuid)
		return nil, result.Err()
	}

	return &ClienteSaldoT{
		Saldo: saldoAtual,
		Cliente: ClienteT{
			ID:     clienteID,
			Limite: limite,
		},
	}, nil
}
