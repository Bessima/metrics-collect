package main

import (
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/logger"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

func main() {
	config := InitConfig()

	if err := run(config.Address); err != nil {
		panic(err)
	}
}

func getMetricRouter(storage *repository.MemStorage, templates *template.Template) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)

	router.Get("/", handler.MainHandler(storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(storage))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(storage))

	return router
}

func run(address string) error {
	templates := handler.ParseAllTemplates()

	storage := repository.NewMemStorage()

	if err := logger.Initialize("debug"); err != nil {
		return err
	}

	logger.Log.Info("Running server on", zap.String("address", address))

	return http.ListenAndServe(address, getMetricRouter(&storage, templates))
}
