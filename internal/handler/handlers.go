package handler

import (
	"github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

func SetMetricHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		typeMetric := chi.URLParam(request, "typeMetric")
		metric := chi.URLParam(request, "name")

		switch repository.TypeMetric(typeMetric) {
		case repository.TypeCounter:
			value, err := strconv.ParseInt(chi.URLParam(request, "value"), 10, 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.Counter(metric, value)
			log.Println("Successful counter: ", metric, storage.View(models.Counter, metric))
		case repository.TypeGauge:
			value, err := strconv.ParseFloat(chi.URLParam(request, "value"), 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			storage.ReplaceGaugeMetric(metric, value)
			log.Println("Successful replacing gauge: ", metric, storage.View(models.Gauge, metric))
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
