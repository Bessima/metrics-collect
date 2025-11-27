package main

import (
	"context"
	configApp "github.com/Bessima/metrics-collect/internal/config"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"go.uber.org/zap"
	"time"
)

type App struct {
	config            *configApp.Config
	storageRepository repository.StorageRepositoryI
	metricsFromFile   *repository.MetricsFromFile
	rootContext       context.Context
}

func NewApp(ctx context.Context, config *configApp.Config, storage repository.StorageRepositoryI) *App {
	app := &App{rootContext: ctx, config: config, storageRepository: storage}

	app.metricsFromFile = repository.NewMetricsFromFile(app.config.FileStoragePath)

	return app
}

func (app *App) loadMetricsFromFile() {
	if app.metricsFromFile == nil {
		return
	}
	switch app.storageRepository.(type) {
	case *repository.FileStorageRepository:
		return
	default:

		if err := app.metricsFromFile.Load(); err != nil {
			logger.Log.Warn(err.Error())
			return
		}
		logger.Log.Info("Metrics was loaded from file", zap.String("path", app.config.FileStoragePath))

		app.storageRepository.Load(app.metricsFromFile.GetMetrics())
	}
}

func (app *App) saveMetricsInFile(ctx context.Context) {
	if app.config.StoreInterval <= 0 {
		return
	}
	if app.metricsFromFile == nil {
		return
	}

	ticker := time.NewTicker(time.Duration(app.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repository.UpdateMetricInFile(app.storageRepository, app.metricsFromFile)
		case <-ctx.Done():
			logger.Log.Info("Stopping metrics saver")
			return
		}
	}
}
