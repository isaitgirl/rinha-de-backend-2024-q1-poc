package rinha

import (
	"context"
	"fmt"
	"os"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/db"
	"rinha-be-2-app/internal/influxdb2"
	"rinha-be-2-app/internal/logger"
	"rinha-be-2-app/internal/metric"
	"time"
)

// Essa rotina de boostrap foi feita só para funcionar, sem nenhum capricho mesmo.

func Bootstrap() {

	logger.Info("Realizando inicialização de dados da aplicação")
	logger.Info("Removendo dados de Clientes - se houver")

	var err error

	query := `drop table if exists clientes;`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao remover dados de clientes no db", "error", err)
		os.Exit(1)
	}

	query = `create table clientes (
		id integer,
		limite integer,
		saldo_inicial integer
		);
	`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao criar tabela de clientes no db", "error", err)
		os.Exit(1)
	}

	query = `insert into clientes (id, limite, saldo_inicial) values
		(1, 100000, 0),
		(2, 80000, 0),
		(3, 1000000, 0),
		(4, 10000000, 0),
		(5, 500000, 0);`

	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao inserir dados de clientes no db", "error", err)
		os.Exit(1)
	}

	err = influxdb2.InfluxDB2Service.Deleter.DeleteWithName(context.TODO(),
		config.AppConfig.InfluxDB2Org,
		config.AppConfig.InfluxDB2Bucket,
		time.Now().In(time.UTC).Add(time.Hour*24*7*-1),
		time.Now(),
		`_measurement="clientes"`,
	)
	if err != nil {
		logger.Warn("Erro ao remover dados de clientes", "error", err)
		os.Exit(1)
	}

	defer logger.Info("Inicialização finalizada com sucesso")

	logger.Info("Inserindo dados iniciais de Clientes")

	// Cria estrutura de transacoes
	query = `drop table if exists transacoes;`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao dropar tabela de transações", "error", err)
		os.Exit(1)
	}

	query = `create unlogged table transacoes (
		cliente smallint,
		valor integer,
		tipo char(1),
		descricao varchar(255),
		realizada_em timestamp);
		`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao criar tabela de transações", "error", err)
		os.Exit(1)
	}

	query = `drop index if exists transacoesByClienteTipo;`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao dropar índice de transações", "error", err)
		os.Exit(1)
	}

	query = `create index transacoesByClienteTipo on transacoes (cliente,tipo);`
	_, err = db.DBService.Client.Exec(query)
	if err != nil {
		logger.Warn("Erro ao criar índice de transações", "error", err)
		os.Exit(1)
	}

	// Escreve no InfluxDB2 os dados iniciais de clientes

	// clientes,cliente=1 saldo_inicial=0,limite=100000
	// clientes,cliente=2 saldo_inicial=0,limite=80000
	// clientes,cliente=3 saldo_inicial=0,limite=1000000
	// clientes,cliente=4 saldo_inicial=0,limite=10000000
	// clientes,cliente=5 saldo_inicial=0,limite=500000

	clientesDatapoints := []metric.DataPointT{
		ClienteT{
			ID:           "1",
			Limite:       100000,
			SaldoInicial: 0,
			Metrica:      "clientes",
		},
		ClienteT{
			ID:           "2",
			Limite:       80000,
			SaldoInicial: 0,
			Metrica:      "clientes",
		},
		ClienteT{
			ID:           "3",
			Limite:       1000000,
			SaldoInicial: 0,
			Metrica:      "clientes",
		},
		ClienteT{
			ID:           "4",
			Limite:       10000000,
			SaldoInicial: 0,
			Metrica:      "clientes",
		},
		ClienteT{
			ID:           "5",
			Limite:       500000,
			SaldoInicial: 0,
			Metrica:      "clientes",
		},
	}

	influxdb2.Write(config.AppConfig.InfluxDB2Org, config.AppConfig.InfluxDB2Bucket, clientesDatapoints...)

	NewClienteService()

	ClienteService.ClientesCache = make(map[string]ClienteT, 5)
	for i, cliente := range clientesDatapoints {
		ClienteService.ClientesCache[clientesDatapoints[i].(ClienteT).ID] = cliente.(ClienteT)
	}

	logger.Info("Novo cache de clientes criado", "mapa", fmt.Sprintf("%+v", ClienteService.ClientesCache))

	logger.Info("Inserindo abertura de contas")

	txTime := time.Now()
	influxdb2.Write(config.AppConfig.InfluxDB2Org, config.AppConfig.InfluxDB2Bucket,
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "c",
			Cliente:   "1",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "d",
			Cliente:   "1",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "c",
			Cliente:   "2",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "d",
			Cliente:   "2",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "c",
			Cliente:   "3",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "d",
			Cliente:   "3",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "c",
			Cliente:   "4",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "d",
			Cliente:   "4",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "c",
			Cliente:   "5",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
		TransacaoEventoT{
			Timestamp: txTime,
			Tipo:      "d",
			Cliente:   "5",
			Descricao: "Abertura de conta",
			Valor:     0,
			Metrica:   "transacoes",
		},
	)

	influxdb2.InfluxDB2Service.Writer.Flush(context.TODO())

}
