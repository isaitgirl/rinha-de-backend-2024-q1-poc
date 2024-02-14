package rinha

import (
	"context"
	"rinha-be-2-app/internal/db"
	"rinha-be-2-app/internal/influxdb2"
	"rinha-be-2-app/internal/logger"
	"rinha-be-2-app/internal/metric"
	"sync"
	"time"
)

type ClienteT struct {
	ID           string
	Limite       int64
	SaldoInicial int64
	Metrica      string
}

type ClienteSaldoT struct {
	*sync.RWMutex
	Cliente ClienteT
	Saldo   int64
}

func (c ClienteT) GetDataPoint() (string, metric.TagT, metric.FieldT, time.Time) {
	t := make(metric.TagT, 0)
	t["cliente"] = c.ID

	f := make(metric.FieldT, 0)
	f["limite"] = c.Limite
	f["saldo_inicial"] = c.SaldoInicial

	return c.Metrica, t, f, time.Now()
}

func getClientKey(id string) string {
	return "clientes/"
}

type ClienteServiceT struct {
	ClientesCache      map[string]ClienteT
	ClientesSaldoCache map[string]ClienteSaldoT
}

var ClienteService *ClienteServiceT

func NewClienteService() {
	ClienteService = &ClienteServiceT{
		ClientesCache:      make(map[string]ClienteT, 5),
		ClientesSaldoCache: make(map[string]ClienteSaldoT, 5),
	}
}

// ================================================================================================================

func ExisteClienteCache(ctx context.Context) bool {

	clienteID := ctx.Value("clienteID").(string)
	uuid := ctx.Value("uuid").(string)

	txStart := time.Now()
	defer func() {
		logger.Info("Consulta de cliente finalizada", "cliente", clienteID, "duracao", time.Since(txStart), "uuid", uuid)
	}()

	logger.Info("Consultando se o cliente existe", "cliente", clienteID, "uuid", uuid)

	_, ok := ClienteService.ClientesCache[ctx.Value("clienteID").(string)]

	switch ok {
	case true:
		logger.Info("Cliente encontrado", "cliente", clienteID, "uuid", uuid)
		return true
	default:
		logger.Info("Cliente inexistente", "cliente", clienteID, "uuid", uuid)
		return false
	}

}

func ExisteCliente(ctx context.Context) bool {

	clienteID := ctx.Value("clienteID").(string)
	uuid := ctx.Value("uuid").(string)

	txStart := time.Now()
	defer func() {
		logger.Info("Consulta de cliente finalizada", "cliente", clienteID, "duracao", time.Since(txStart), "uuid", uuid)
	}()

	logger.Info("Consultando se o cliente existe", "cliente", clienteID, "uuid", uuid)

	query := `select 1 from clientes where id = '` + clienteID + `'`
	var retVal int

	err := db.DBService.Client.QueryRow(query).Scan(&retVal)

	if err != nil {
		//spew.Dump(err)
		//logger.Warn("Erro ao consultar cliente", "erro", err, "uuid", uuid)
		return false
	}

	return true

}

func ExisteClienteFlux(ctx context.Context) bool {

	clienteID := ctx.Value("clienteID").(string)
	uuid := ctx.Value("uuid").(string)

	txStart := time.Now()
	defer func() {
		logger.Info("Consulta de cliente finalizada", "cliente", clienteID, "duracao", time.Since(txStart), "uuid", uuid)
	}()

	logger.Info("Consultando se o cliente existe", "cliente", clienteID, "uuid", uuid)

	query := `from(bucket: "rinha")
						|> range(start: -7d, stop: now())
						|> filter(fn: (r) => r["_measurement"] == "clientes")
						|> filter(fn: (r) => r["cliente"] == "` + clienteID + `")
						|> filter(fn: (r) => r["_field"] == "saldo_inicial")
						|> keep(columns: ["_value"])
						|> count()
	`

	result, err := influxdb2.Query(query)
	if err != nil {
		logger.Warn("Erro ao consultar cliente", "erro", err, "uuid", uuid)
		return false
	}

	for result.Next() {
		logger.Info("Cliente encontrado", "cliente", clienteID, "uuid", uuid)
		return true
	}

	if result.Err() != nil {
		logger.Warn("Erro ao iterar consulta de cliente", "erro", result.Err(), "uuid", uuid)
		return false
	}

	logger.Info("Cliente inexistente", "cliente", clienteID, "uuid", uuid)
	return false
}
