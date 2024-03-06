package main

import (
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/deleteuserurl"
	grpc2 "github.com/gerasimovpavel/shortener.git/internal/grpc"
	shortener "github.com/gerasimovpavel/shortener.git/internal/proto"
	"github.com/gerasimovpavel/shortener.git/internal/router"
	"github.com/gerasimovpavel/shortener.git/internal/service"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
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
	defer logger.Logger.Sync()
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

	signalErr := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var wg sync.WaitGroup

	go func() {
		sig := <-signalCh
		logger.Logger.Info(fmt.Sprintf("Signal %v recieved. Shutdowning...\n", sig))
		wg.Wait()
		os.Exit(0)
	}()

	logger.Logger.Info("starting server", zap.String("address", config.Cfg.Host))

	switch config.Cfg.SSLEnabled {
	case true:
		server.ListenAndServeTLS(config.Cfg.SSLCert, config.Cfg.SSLKey)
	case false:

		server.ListenAndServe()
	}

	service := service.NewService(&config.Cfg, storage.Stor, logger.Logger)

	if config.Cfg.GRPCPort != "" {
		wg.Add(1)
		go func(errs chan<- error) {
			defer wg.Done()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Cfg.GRPCPort))
			if err != nil {
				logger.Logger.Sugar().Errorf("failed to listen: %w", err)
				errs <- err
				return
			}
			grpcServer := grpc.NewServer()
			reflection.Register(grpcServer)
			shortener.RegisterShortenerServer(grpcServer, grpc2.NewGRPCService(logger.Logger.Sugar(), service))
			logger.Logger.Sugar().Infof("running gRPC service on %s", config.Cfg.GRPCPort)
			if err = grpcServer.Serve(lis); err != nil {
				if errors.Is(err, grpc.ErrServerStopped) {
					return
				}
				errs <- err
			}
		}(signalErr)
	}

}
