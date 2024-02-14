package rinha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rinha-be-2-app/internal/logger"
	"rinha-be-2-app/internal/rpc"
	"strings"
	"time"

	"github.com/google/uuid"
)

func NovaRequisicao(w http.ResponseWriter, r *http.Request) {

	logger.Info("NovaRequisicao", "method", r.Method, "path", r.URL.Path)

	fullPath := r.URL.Path
	splitedPath := strings.Split(fullPath[1:], "/")

	logger.Info("NovaRequisicao", "splitedPath", splitedPath)

	if len(splitedPath) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("path inválido, esperado /clientes/{id}/{operacao} (transacoes ou extrato)"))
		return
	}
	clienteID := splitedPath[1]
	operacao := splitedPath[2]

	ctx, ctxCancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "clienteID", clienteID)
	ctx = context.WithValue(ctx, "txnStart", time.Now())
	ctx = context.WithValue(ctx, "uuid", uuid.New().String())

	go func() {
		select {
		case <-r.Context().Done():
			ctxCancel()
		}
	}()

	if !ExisteCliente(ctx) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch operacao {
	case "transacoes":
		NovaTransacao(ctx, w, r)
	case "extrato":
		ObterExtrato(ctx, w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("operacao inválida, esperado transacoes ou extrato"))
		return
	}

	return

}

func NovaTransacao(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	clienteID := ctx.Value("clienteID").(string)
	if clienteID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("clienteID não informado"))
		return
	}

	txStart := ctx.Value("txnStart").(time.Time)
	defer func() {
		logger.Info("Requisição de NovaTransação finalizada", "clienteID", clienteID, "duracao", time.Since(txStart), "uuid", ctx.Value("uuid").(string))
	}()

	logger.Info("Requisição para NovaTransacao", "method", r.Method, "path", r.URL.Path, "clienteID", clienteID, "uuid", ctx.Value("uuid").(string))

	logger.Info("Lendo o body da requisição", "clienteID", clienteID, "uuid", ctx.Value("uuid").(string))

	txBodyReadStart := time.Now()
	// Lê os bytes do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("falha ao ler o body da requisição"))
		return
	}

	logger.Info("Body lido com sucesso", "clienteID", clienteID, "duracao", time.Since(txBodyReadStart), "uuid", ctx.Value("uuid").(string))

	logger.Info("Deserializando o JSON", "clienteID", clienteID, "uuid", ctx.Value("uuid").(string))
	txUnmarshalStart := time.Now()

	// Tenta deserializar o JSON para um objeto TransacaoT
	var transacao TransacaoT
	err = json.Unmarshal(body, &transacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Info("JSON deserializado com sucesso", "clienteID", ctx.Value("clienteID").(string), "duracao", time.Since(txUnmarshalStart), "uuid", ctx.Value("uuid").(string))

	clienteSaldo, err := ExecutaTransacao(ctx, transacao)
	if err != nil {
		logger.Warn("Erro ao executar a transação", "clienteID", ctx.Value("clienteID").(string), "erro", err.Error(), "uuid", ctx.Value("uuid").(string))
		var erroLimite *TransacaoErroLimiteUltrapassadoT
		var erroInvalido *TransacaoErroTipoInvalidoT
		switch {
		case errors.As(err, &erroLimite):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.As(err, &erroInvalido):
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	logger.Info("Novo saldo do cliente", "clienteID", ctx.Value("clienteID").(string), "saldo", clienteSaldo.Saldo, "uuid", ctx.Value("uuid").(string))
	logger.Info("Requisição de NovaTransacao executada com sucesso", "clienteID", ctx.Value("clienteID").(string), "uuid", ctx.Value("uuid").(string))

	saldoResp := struct {
		Limite int64 `json:"limite"`
		Saldo  int64 `json:"saldo"`
	}{
		Limite: clienteSaldo.Cliente.Limite,
		Saldo:  clienteSaldo.Saldo,
	}

	saldoClienteBytes, err := json.Marshal(saldoResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("falha ao serializar o saldo do cliente"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(saldoClienteBytes)

	return
}

func ObterExtrato(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	clienteID := ctx.Value("clienteID").(string)
	if clienteID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("clienteID não informado"))
		return
	}

	txStart := ctx.Value("txnStart").(time.Time)
	defer func() {
		logger.Info("Requisição de ObterExtrato finalizada", "clienteID", clienteID, "duracao", time.Since(txStart), "uuid", ctx.Value("uuid").(string))
	}()

	logger.Info("Requisição para ObterExtrato", "method", r.Method, "path", r.URL.Path, "clienteID", ctx.Value("clienteID").(string), "uuid", ctx.Value("uuid").(string))

	extrato, err := MontaExtrato(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(extrato)
	if err != nil {
		logger.Warn("Erro ao serializar o extrato", "clienteID", clienteID, "erro", err, "uuid", ctx.Value("uuid").(string))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	return
}

func LockAcquire(w http.ResponseWriter, r *http.Request) {

	logger.Info("LockAcquire", "method", r.Method, "path", r.URL.Path)
	semaphore := &rpc.SemaphoreT{}
	err := rpc.ClientService.Acquire("app1", semaphore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erro ao adquirir lock:" + err.Error()))
		return
	}

	logger.Info("LockAcquired", "semaphore", fmt.Sprintf("%+v", semaphore))

	semaphoreBytes, err := json.Marshal(semaphore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erro ao serializar o semáforo"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(semaphoreBytes)

	return
}

func LockRelease(w http.ResponseWriter, r *http.Request) {
	logger.Info("LockRelease", "method", r.Method, "path", r.URL.Path)
	semaphore := &rpc.SemaphoreT{}
	err := rpc.ClientService.Release("app1", semaphore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erro ao liberar lock:" + err.Error()))
		return
	}

	logger.Info("LockRelease", "semaphore", fmt.Sprintf("%+v", semaphore))

	semaphoreBytes, err := json.Marshal(semaphore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("erro ao serializar o semáforo"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(semaphoreBytes)
	return
}
