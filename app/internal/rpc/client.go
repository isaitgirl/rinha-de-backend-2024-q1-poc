package rpc

import (
	"net/rpc"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/logger"
	"time"
)

type ClientServiceT struct {
	Client *rpc.Client
}

var ClientService *ClientServiceT

func NewClientService() {

	var err error
	ClientService = &ClientServiceT{}
	remotePeer := config.AppConfig.RemotePeerRPCAddr

	logger.Info("Tentando conectar ao RPC remoto", "remotePeer", remotePeer)
	for {
		ClientService.Client, err = rpc.Dial("tcp", remotePeer)
		if err != nil {
			logger.Warn("Falha ao conectar ao RPC remoto", "remotePeer", remotePeer, "error", err.Error())
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	logger.Info("Cliente RPC iniciado e conectado")

}

func (c *ClientServiceT) HealthCheck(app string, reply *SemaphoreT) error {
	for {
		time.Sleep(5 * time.Second)
		err := c.Client.Call("RPCServer.Ping", true, &reply)
		if err != nil {
			logger.Warn("Falha ao verificar conexão com RPC remoto", "error", err.Error())
		}
	}
}

func (c *ClientServiceT) Acquire(app string, reply *SemaphoreT) error {
	err := c.Client.Call("RPCServer.Acquire", app, &reply)
	return err
}

func (c *ClientServiceT) Release(app string, reply *SemaphoreT) error {
	err := c.Client.Call("RPCServer.Release", app, &reply)
	return err
}
