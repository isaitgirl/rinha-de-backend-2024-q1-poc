package logger

import (
	"fmt"
	"log"
	"os"
	"rinha-be-2-app/internal/config"
)

// Logger é o singleton que dá acesso às funcionalidades do pacote logger
var Logger *log.Logger

// Info é um wrapper para logs do tipo INFO
func Info(msg string, pairs ...interface{}) {
	Logger.Println("INFO", buildLogKeysAndValues(msg, pairs...))
}

// Warn é um wrapper para logs do tipo WARN
func Warn(msg string, pairs ...interface{}) {
	Logger.Println("WARN", buildLogKeysAndValues(msg, pairs...))
}

// Error é um wrapper para logs do tipo ERROR
func Error(msg string, pairs ...interface{}) {
	Logger.Println("ERROR", buildLogKeysAndValues(msg, pairs...))
}

// Debug é um wrapper para logs do tipo DEBUG
func Debug(msg string, pairs ...interface{}) {
	Logger.Println("DEBUG", buildLogKeysAndValues(msg, pairs...))
}

// Monta a estrutura de logs com modelo logfmt (chave-valor)
func buildLogKeysAndValues(msg string, pairs ...interface{}) []interface{} {
	keysAndValues := make([]interface{}, 0)
	keysAndValues = append(keysAndValues, parseBeforePrint("msg", msg))
	keysAndValues = append(keysAndValues, parseBeforePrint(pairs...))
	return keysAndValues
}

// Avalia os pares recebidos, se for o primeiro, não adiciona o prefixo pois é a msg em si
func parseBeforePrint(r ...interface{}) (v string) {
	for i, e := range r {
		if i%2 == 0 {
			v += fmt.Sprintf(" %v=", e)
		} else {
			v += fmt.Sprintf("%v", e)
		}
	}
	return
}

// NewLogger instancia o single para acesso ao logger
func NewLogger() {
	logFile, err := os.OpenFile(config.AppConfig.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
		panic("Failed to open log file")
	}
	Logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)

}
