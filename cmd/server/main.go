package main

import (
	"github.com/Bessima/metrics-collect/internal/handler"
	"github.com/Bessima/metrics-collect/internal/repository"
	"net/http"
)

var STORAGE repository.MemStorage

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	STORAGE = repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/{typeMetric}/{name}/{value}", handler.SetMetricHandler(&STORAGE))
	return http.ListenAndServe(`:8080`, mux)
}
