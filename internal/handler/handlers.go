package handler

import (
	"github.com/Bessima/metrics-collect/internal/common"
	"github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strconv"
)

func SetMetricHandler(storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) http.HandlerFunc {
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
			newValue, _ := storage.GetValue(models.Counter, metric)
			log.Println("Successful counter: ", metric, newValue)
		case repository.TypeGauge:
			value, err := strconv.ParseFloat(chi.URLParam(request, "value"), 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			storage.ReplaceGaugeMetric(metric, value)
			newValue, _ := storage.GetValue(models.Gauge, metric)
			log.Println("Successful replacing gauge: ", metric, newValue)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if metricsFromFile != nil {
			repository.UpdateMetricInFile(storage, metricsFromFile)
		}
	}
}

func ViewMetricValue(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		typeMetric := repository.TypeMetric(chi.URLParam(request, "typeMetric"))
		metric := chi.URLParam(request, "name")
		value, err := storage.GetValue(typeMetric, metric)
		if err != nil {
			log.Println("Failed to view metric value, error: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		result, err := common.ConvertInterfaceToStr(value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, result)
	}
}
