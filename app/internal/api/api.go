package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"rinha-be-2-app/internal/logger"
)

var (
	//HTTPService provê acesso às funcionalidades HTTP da aplicação
	HTTPService *HTTPServiceT
	routeMux    *http.ServeMux
	endpoints   map[string]func(http.ResponseWriter, *http.Request)
)

// HTTPServiceT define a estrutura de serviço para HTTPService
type HTTPServiceT struct{}

// NewHTTPService inicializa o singleton que permite acesso às funções de serviço HTTP
func NewHTTPService() {
	HTTPService = &HTTPServiceT{}
}

// AddEndpoint método que permite adicionar endpoints ao serviço padrão HTTP
func (h *HTTPServiceT) AddEndpoint(endpoint string, f func(http.ResponseWriter, *http.Request)) {
	routeMux.HandleFunc(endpoint, f)
}

// StartServer inicia um novo endpoint HTTP para a aplicação
func (h *HTTPServiceT) StartServer(addr, port string) {

	// Inicia um HTTP listener para responder health checks e métricas do Prometheus
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport.DisableKeepAlives = true
		transport.MaxIdleConnsPerHost = 1
		transport.CloseIdleConnections()
	}

	// Adiciona as rotas padrão do serviço HTTP
	// Para adicionar endpoint extras, `AddEndpoint` deve ser invocado pelos pacotes externos ANTES do StartServer
	h.AddEndpoint("/ping", ping)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      routeMux,
	}

	logger.Info("Starting HTTP server", "addr", addr, "port", port)
	err := server.ListenAndServe()
	if err != nil {
		logger.Error("Unable to open a listener",
			"errorDesc", err.Error(),
		)
		os.Exit(1)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`pong`))
	return
}

func init() {
	routeMux = http.NewServeMux()
}
