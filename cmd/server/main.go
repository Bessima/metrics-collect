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
	flags := ServerFlags{}
	flags.Init()

	log.Println("Running server on", flags.address)

	if err := run(flags.address); err != nil {
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
