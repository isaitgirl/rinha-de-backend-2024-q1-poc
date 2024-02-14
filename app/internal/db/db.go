package db

import (
	"database/sql"
	"fmt"
	"os"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/logger"

	_ "github.com/lib/pq"
)

type DBServiceT struct {
	Client *sql.DB
}

var DBService *DBServiceT

func NewDBService() {

	logger.Info("Inicializando conexão com banco de dados")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo client_encoding=UTF8 password=%s",
		config.AppConfig.TimescaleDBHost,
		config.AppConfig.TimescaleDBUser,
		config.AppConfig.TimescaleDBCatalog,
		config.AppConfig.TimescaleDBPort,
		config.AppConfig.TimescaleDBPass,
	)
	sqlConn, err := sql.Open("postgres", dsn)

	DBService = &DBServiceT{
		Client: sqlConn,
	}

	if err != nil {
		logger.Warn("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	err = DBService.Client.Ping()
	if err != nil {
		logger.Warn("Failed to ping database", "error", err)
		os.Exit(1)
	}

	logger.Info("Conectado ao banco de dados com sucesso")
}
