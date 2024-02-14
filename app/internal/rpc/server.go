package rpc

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/logger"
	"sync"
	"time"
)

// RPCServer é o tipo que implementa o serviço RPC
type RPCServer int

var SemaphoreLock = &sync.Mutex{}

type SemaphoreT struct {
	AcquiredAt time.Time
	AcquiredBy string
	Active     bool
}

var Semaphore = &SemaphoreT{
	time.Now(),
	"",
	false,
}

type AcquireTimeoutErrorT struct {
	Message string
}

func (e AcquireTimeoutErrorT) Error() string {
	return e.Message
}

var AcquireTimeoutError = AcquireTimeoutErrorT{Message: "Tempo excedido ao tentar adquirir o lock"}

// NewRPCServer inicia o serviço RPC
func NewRPCServer() {
	server := rpc.NewServer()
	server.Register(new(RPCServer))
	rpcPortStr := fmt.Sprintf("%s:%s", config.AppConfig.RPCServerHost, config.AppConfig.RPCServerPort)
	l, err := net.Listen("tcp", rpcPortStr)

	if err != nil {
		logger.Warn("Falha ao iniciar servidor RPC", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Servidor RPC iniciado ", "rpc", rpcPortStr)

	go server.Accept(l)
}

// Ping permite verificar se o serviço está ativo
func (r *RPCServer) Ping(arg bool, res *bool) error {
	*res = true
	return nil
}

// Acquire permite adquirir um lock
func (r *RPCServer) Acquire(instance string, res *SemaphoreT) error {

	txStart := time.Now().In(time.Local)

	logger.Info("Tentando obter lock do semáforo", "instancia", instance, "iniciado_em", txStart)
	SemaphoreLock.Lock()

	Semaphore.AcquiredAt = time.Now()
	Semaphore.AcquiredBy = instance
	Semaphore.Active = true

	elapsed := int(time.Since(txStart))
	logger.Info("Lock do semáforo obtido", "instancia", instance, "duracao_em_ns", elapsed)
	res.AcquiredBy = instance
	res.AcquiredAt = time.Now()
	res.Active = true

	return nil
}

func (r *RPCServer) Release(instance string, res *SemaphoreT) error {
	SemaphoreLock.Unlock()
	Semaphore.Active = false
	Semaphore.AcquiredBy = ""
	res.AcquiredBy = ""
	res.AcquiredAt = time.Time{}
	res.Active = false
	logger.Info("Liberação do lock do semáforo realizada", "instancia", instance)
	return nil
}
