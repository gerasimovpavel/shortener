package main

import (
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/logger"
	"github.com/gerasimovpavel/shortener.git/internal/router"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"net/http"
	_ "net/http/pprof"
)

// main
func main() {
	var err error
	//Logger
	err = logger.NewLogger()
	if err != nil {
		panic(err)
	}
	//Парсим переменные и аргументы команднй строки
	config.ParseEnvFlags()
	// создаем Storage
	storage.Stor, err = storage.NewStorage()
	if err != nil {
		panic(err)
	}
	// URLDeleter
	deleteuserurl.URLDel = deleteuserurl.NewURLDeleter()
	// запускаем сервер
	router := router.MainRouter()
	if router == nil {
		panic(errors.New("failed to create main router"))
	}
	err = http.ListenAndServe(config.Options.Host, router)
	if err != nil {
		panic(err)
	}
}
