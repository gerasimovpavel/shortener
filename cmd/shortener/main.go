package main

import (
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/handlers"
	"net/http"
)

// main
func main() {
	//Парсим переменные и аргументы команднй строки
	config.ParseEnvFlags()
	// запускаем сервер
	err := http.ListenAndServe(config.Options.Host, handlers.MainRouter())
	if err != nil {
		panic(err)
	}
}
