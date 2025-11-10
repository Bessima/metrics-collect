package main

import (
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/middlewares/compress"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"time"
)

func main() {

	if err := run(); err != nil {
		panic(err)
	}
}

func getMetricRouter(storage *repository.MemStorage, templates *template.Template, metricsFromFile *repository.MetricsFromFile) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GZIPMiddleware)

	router.Get("/", handler.MainHandler(storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(storage, metricsFromFile))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(storage))
	router.Post("/update/", handler.UpdateHandler(storage, metricsFromFile))
	router.Post("/value/", handler.ValueHandler(storage))

	return router
}

func run() error {
	if err := logger.Initialize("debug"); err != nil {
		return err
	}

	templates := handler.ParseAllTemplates()
	config := InitConfig()
	storage := repository.NewMemStorage()

	metricsFromFile := repository.MetricsFromFile{FileName: *config.FileStoragePath}

	if *config.Restore {
		if err := metricsFromFile.Load(); err != nil {
			logger.Log.Warn(err.Error())
		} else {
			logger.Log.Info("Metrics was loaded from file", zap.String("path", *config.FileStoragePath))
			storage.Load(metricsFromFile.GetMetrics())
		}
	}

	logger.Log.Info("Running server on", zap.String("address", *config.Address))

	go func() {
		if *config.StoreInterval <= 0 {
			return
		}
		for {
			repository.UpdateMetricInFile(&storage, &metricsFromFile)
			time.Sleep(time.Duration(*config.StoreInterval) * time.Second)
		}
	}()

	if *config.StoreInterval == 0 {
		return http.ListenAndServe(*config.Address, getMetricRouter(&storage, templates, &metricsFromFile))
	}

	return http.ListenAndServe(*config.Address, getMetricRouter(&storage, templates, nil))
}
