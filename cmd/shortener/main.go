package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	"github.com/gerasimovpavel/shortener.git/internal/router"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	server := &http.Server{
		Addr:              config.Cfg.Host,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	idleConnsClosed := make(chan struct{})

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		<-sigint

		logger.Logger.Info("shutting down server...")

		if err := server.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("HTTP server Shutdown: %v", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	logger.Logger.Info("starting server", zap.String("address", config.Cfg.Host))

	switch config.Cfg.SSLEnabled {
	case true:
		logger.Logger.Fatal(
			"https server down",
			zap.Error(server.ListenAndServeTLS(config.Cfg.SSLCert, config.Cfg.SSLKey)),
		)

	case false:
		logger.Logger.Fatal("http server down", zap.Error(server.ListenAndServe()))
	}

	<-idleConnsClosed

	wg.Wait()

}
