package app

import (
	"context"
	"os"
	"os/signal"
	"rinha-be-2-app/internal/api"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/db"
	"rinha-be-2-app/internal/influxdb2"
	"rinha-be-2-app/internal/logger"
	"rinha-be-2-app/internal/rinha"
	"rinha-be-2-app/internal/rpc"

	"syscall"
)

var (
	MainContext context.Context
	MainCancel  context.CancelFunc
)

// New inicializa o singleton para acesso às funcionalidades da aplicação
func New() {

	MainContext, MainCancel = context.WithCancel(context.Background())

	// =========================================================================================

	// Lança a goroutine que captura os sinais de desligamento
	go terminationListener()

	// Inicializa o singleton para acesso às funções de serviço HTTP padrão
	//api.NewHTTPService()

	// Inicia o singleton de acesso às configurações
	config.NewConfig()

	// Inicia o singleton de acesso ao DLogger
	logger.NewLogger()

	logger.Info("Iniciando aplicação...")

	// Inicia o singleton de acesso às métricas
	//metric.NewMetricsService(MetricsPrefix)

	// =========================================================================================

	// Inicialização da lógica da aplicação

	influxdb2.NewInfluxDB2Service()

	db.NewDBService()

	rpc.NewRPCServer()
	rpc.NewClientService()

	// Escreve os parâmetros iniciais no InfluxDB2
	rinha.Bootstrap()

	// =========================================================================================
	// Define as rotas, seus handlers e inicia o serviço HTTP padrão

	api.HTTPService.AddEndpoint("/", rinha.NovaRequisicao)
	api.HTTPService.AddEndpoint("/lock", rinha.LockAcquire)
	api.HTTPService.AddEndpoint("/release", rinha.LockRelease)
	api.HTTPService.StartServer(config.AppConfig.HTTPHost, config.AppConfig.HTTPPort)

}

// Aguarda os variados sinais de terminação para iniciar o shutdown limpo da aplicação
func terminationListener() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT, syscall.SIGHUP)

	for {
		sig := <-c
		if sig != nil {

			// TODO: Se receber um SIGHUP, recarregar as configurações. Se conseguir, restarta a Cron da aplicação
			if sig == syscall.SIGHUP {

				logger.Info("SIGHUP signal received, performing configuration reload")
				//logger.Reload()
				continue
			}

			MainCancel()
			logger.Info("Shutdown signal received, waiting context do finish")

			// Encerra as queues
			//queue.PQ.Close()
			//queue.DQ.Close()

			// TODO: implementar um timer máximo para shutdown
			//<-MainContext.Done()

			influxdb2.InfluxDB2Service.Client.Close()

			logger.Info("Ready to shutdown")
			os.Exit(0)
		}
	}
}
