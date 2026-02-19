package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Bessima/metrics-collect/internal/config"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/internal/service"
	"github.com/Bessima/metrics-collect/pkg/audit"
	"go.uber.org/zap"
)

func main() {
	err := initLogger()
	if err != nil {
		logger.Log.Warn(err.Error())
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf := config.InitConfig()
	storageService := service.NewStorageService(rootCtx, conf)
	defer storageService.Close()

	storageService.GetRepository()

	app := NewApp(rootCtx, conf, *storageService.GetRepository())

	if app.config.Restore {
		app.loadMetricsFromFile()
	}

	event := audit.Event{}
	if app.config.AuditFile != "" || app.config.AuditURL != "" {

		if app.config.AuditFile != "" {
			fileSubscriber := audit.NewFileSubscriber(app.config.AuditFile)
			event.Register(fileSubscriber)
		}
		if app.config.AuditURL != "" {
			urlSubscriber := audit.NewURLSubscriber(app.config.AuditURL)
			event.Register(urlSubscriber)
		}

	}

	serverService := service.NewServerService(rootCtx, conf.Address, conf.KeyHash, app.storageRepository)
	serverService.SetRouter(conf.StoreInterval, app.metricsFromFile, &event)

	saveCtx, saveCancel := context.WithCancel(rootCtx)
	defer saveCancel()
	go app.saveMetricsInFile(saveCtx)

	serverErr := make(chan error, 1)
	logger.Log.Info("Running Server on", zap.String("address", conf.Address))
	go serverService.RunServer(&serverErr)

	// Ждем сигнал завершения или ошибку сервера
	var err error
	select {
	case <-rootCtx.Done():
		logger.Log.Info("Received shutdown signal, shutting down.")
	case err = <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
	}

	if shutdownErr := serverService.Shutdown(); shutdownErr != nil {
		logger.Log.Error("Server shutdown error", zap.Error(shutdownErr))
	}

	logger.Log.Info("Received shutdown signal, shutting down.")
	repository.UpdateMetricInFile(app.storageRepository, app.metricsFromFile)

	saveCancel()

	return err
}

func initLogger() error {
	if err := logger.Initialize("debug"); err != nil {
		return err
	}
	return nil
}
