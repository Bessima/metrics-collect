package main

import (
	"context"
	"github.com/Bessima/metrics-collect/internal/config/db"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"go.uber.org/zap"
	"time"
)

type App struct {
	config          *Config
	storage         *repository.MemStorage
	metricsFromFile repository.MetricsFromFile
	rootContext     context.Context
}

func NewApp(ctx context.Context, storage *repository.MemStorage) *App {
	app := &App{rootContext: ctx}

	app.config = InitConfig()
	if storage != nil {
		app.storage = storage
	} else {
		newStorage := repository.NewMemStorage()
		app.storage = &newStorage
	}
	app.metricsFromFile = repository.MetricsFromFile{FileName: app.config.FileStoragePath}

	return app
}

func (app *App) loadMetricsFromFile() {
	if err := app.metricsFromFile.Load(); err != nil {
		logger.Log.Warn(err.Error())
		return
	}
	logger.Log.Info("Metrics was loaded from file", zap.String("path", app.config.FileStoragePath))
	app.storage.Load(app.metricsFromFile.GetMetrics())
}

func (app *App) initDB() *db.DB {
	dbObj, errDB := db.NewDB(app.rootContext, app.config.DatabaseDNS)

	if errDB != nil {

		logger.Log.Error(
			"Unable to connect to database",
			zap.String("path", app.config.DatabaseDNS),
			zap.String("error", errDB.Error()),
		)
	}

	return dbObj
}

func (app *App) saveMetricsInFile(ctx context.Context) {
	if app.config.StoreInterval <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(app.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			repository.UpdateMetricInFile(app.storage, &app.metricsFromFile)
		case <-ctx.Done():
			logger.Log.Info("Stopping metrics saver")
			return
		}
	}
}
