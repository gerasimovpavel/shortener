package main

import (
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"net/http"
)

// main
func main() {
	//Парсим переменные и аргументы команднй строки
	config.ParseEnvFlags()
	// создаем Storage
	err := storage.NewStorage()
	if err != nil {
		panic(err)
	}
	// запускаем сервер
	router, err := handlers.MainRouter()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(config.Options.Host, router)
	if err != nil {
		panic(err)
	}
}
