package main

import (
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"github.com/gerasimovpavel/shortener.git/internal/log"
	"go.uber.org/zap"
	"net/http"
)

// main
func main() {
	//Парсим переменные и аргументы команднй строки
	config.ParseEnvFlags()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	log.Sugar = *logger.Sugar()

	// запускаем сервер
	err = http.ListenAndServe(config.Options.Host, log.WithLogging(handlers.MainRouter()))
	if err != nil {
		panic(err)
	}
}
