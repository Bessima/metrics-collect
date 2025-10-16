package main

import (
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

var STORAGE repository.MemStorage

func main() {
	parseFlags()
	log.Println("Running server on", flagRunAddr)

	if err := run(); err != nil {
		panic(err)
	}
}

func GetMetricRouter(storage repository.MemStorage) chi.Router {
	router := chi.NewRouter()
	router.Get("/", handler.MainHandler(&storage))
	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(&storage))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(&storage))

	return router
}

func run() error {
	STORAGE = repository.NewMemStorage()

	return http.ListenAndServe(flagRunAddr, GetMetricRouter(STORAGE))
}
