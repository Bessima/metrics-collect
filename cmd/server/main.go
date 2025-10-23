package main

import (
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
)

func main() {
	config := InitConfig()

	log.Println("Running server on", config.Address)

	if err := run(config.Address); err != nil {
		panic(err)
	}
}

func getMetricRouter(storage *repository.MemStorage, templates *template.Template) chi.Router {
	router := chi.NewRouter()

	router.Get("/", handler.MainHandler(storage, templates))

	router.Post("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(storage))
	router.Get("/value/{typeMetric}/{name}", handler.ViewMetricValue(storage))

	return router
}

func run(address string) error {
	templates := handler.ParseAllTemplates()

	storage := repository.NewMemStorage()

	return http.ListenAndServe(address, getMetricRouter(&storage, templates))
}
