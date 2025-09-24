package main

import (
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
	"log"
	"net/http"
	"strconv"

	"github.com/Bessima/metrics-collect/internal/repository"
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
	mux.HandleFunc("/update/{typeMetric}/{metric}/{value}", set)
	return http.ListenAndServe(`:8080`, mux)
}

func set(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	typeMetric := request.PathValue("typeMetric")
	metric := request.PathValue("metric")

	switch repository.TypeMetric(typeMetric) {
	case repository.TypeCounter:
		value, err := strconv.ParseInt(request.PathValue("value"), 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		STORAGE.Counter(metric, value)
		log.Println("Successful counter: ", metric, STORAGE.View(models.Counter, metric))
	case repository.TypeGauge:
		value, err := strconv.ParseFloat(request.PathValue("value"), 10)
		if err != nil {
			log.Fatalf("Failed to parse metric value, error: ", err)
		}

		STORAGE.Replace(metric, value)
		log.Println("Successful replacing gauge: ", metric, STORAGE.View(models.Gauge, metric))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
