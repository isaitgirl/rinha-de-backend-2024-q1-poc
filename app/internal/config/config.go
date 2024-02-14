package config

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/kelseyhightower/envconfig"
)

type AppConfigT struct {
	DryRun             bool   `envconfig:"dryrun" default:"true"`
	Environment        string `envconfig:"environment" default:"dev"`
	LogFile            string `envconfig:"APP_LOG_FILE" default:"/app/log/app.log"`
	LogLevel           string `envconfig:"APP_LOG_LEVEL" default:"info"`
	EtcdHost           string `envconfig:"etcd_host"`
	EtcdPort           string `envconfig:"etcd_port"`
	InfluxDB2Host      string `envconfig:"influxdb2_host"`
	InfluxDB2Port      string `envconfig:"influxdb2_port"`
	InfluxDB2Token     string `envconfig:"influxdb2_token"`
	InfluxDB2Org       string `envconfig:"influxdb2_org"`
	InfluxDB2Bucket    string `envconfig:"influxdb2_bucket"`
	HTTPHost           string `envconfig:"http_host"`
	HTTPPort           string `envconfig:"http_port"`
	TimescaleDBHost    string `envconfig:"timescaledb_host"`
	TimescaleDBPort    string `envconfig:"timescaledb_port"`
	TimescaleDBUser    string `envconfig:"timescaledb_user"`
	TimescaleDBPass    string `envconfig:"timescaledb_pass"`
	TimescaleDBCatalog string `envconfig:"timescaledb_catalog"`
	RPCServerHost      string `envconfig:"rpc_server_host"`      // ex: localhost
	RPCServerPort      string `envconfig:"rpc_server_port"`      // ex: 5555
	RemotePeerRPCAddr  string `envconfig:"remote_peer_rpc_addr"` // ex: localhost:5556
}

var (
	AppConfig AppConfigT
)

// NewConfig inicializa as configurações vindas de variáveis de ambiente
func NewConfig() {

	// Carrega as configurações da aplicação
	// As variáveis de ambiente são no formato APP_*
	err := envconfig.Process("app", &AppConfig)
	spew.Dump(AppConfig)
	if err != nil {
		panic("Failed to load application config from environment")
	}
}
