package main

import (
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/router"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
	"net/http"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main
func main() {
	var err error

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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
	switch config.Options.SSL.Enabled {
	case true:
		{
			err = http.ListenAndServeTLS(
				config.Options.Host,
				config.Options.SSL.Cert,
				config.Options.SSL.Key,
				router)
		}
	default:
		{
			err = http.ListenAndServe(config.Options.Host, router)
		}
	}

	if err != nil {
		panic(err)
	}
}
