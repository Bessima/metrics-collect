package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/internal/service"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
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

	app := NewApp(rootCtx, nil)

	if app.config.Restore {
		app.loadMetricsFromFile()
	}

	db := app.initDB()
	serverService := service.NewServerService(rootCtx, app.config.Address, app.storage)
	serverService.SetRouter(app.config.StoreInterval, db.Pool, &app.metricsFromFile)

	saveCtx, saveCancel := context.WithCancel(rootCtx)
	defer saveCancel()
	go app.saveMetricsInFile(saveCtx)

	serverErr := make(chan error, 1)
	logger.Log.Info("Running Server on", zap.String("address", app.config.Address))
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
	repository.UpdateMetricInFile(app.storage, &app.metricsFromFile)
	defer db.Close()

	saveCancel()

	return err
}

func initLogger() error {
	if err := logger.Initialize("debug"); err != nil {
		return err
	}
	return nil
}
