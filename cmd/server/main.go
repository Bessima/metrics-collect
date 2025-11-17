package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
	"time"
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

	app := NewApp(rootCtx)

	if app.config.Restore {
		app.loadMetricsFromFile()
	}

	db := app.initDB()
	server := app.getServer(db.Pool)

	saveCtx, saveCancel := context.WithCancel(rootCtx)
	defer saveCancel()
	go app.saveMetricsInFile(saveCtx)

	serverErr := make(chan error, 1)
	go func() {
		logger.Log.Info("Running server on", zap.String("address", app.config.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()

	// Ждем сигнал завершения или ошибку сервера
	var err error
	select {
	case <-rootCtx.Done():
		logger.Log.Info("Received shutdown signal, shutting down.")
	case err = <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		logger.Log.Error("Server shutdown error", zap.Error(shutdownErr))
	}

	logger.Log.Info("Received shutdown signal, shutting down.")
	repository.UpdateMetricInFile(&app.storage, &app.metricsFromFile)
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
