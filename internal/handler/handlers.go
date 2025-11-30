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

func SetMetricHandler(storage repository.StorageRepositoryI, metricsFromFile *repository.MetricsFromFile) http.HandlerFunc {
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

			if err = storage.Counter(metric, value); err != nil {
				log.Println("Failed to change delta of counter metric, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			newValue, err := storage.GetValue(models.Counter, metric)
			if err != nil {
				log.Println("Failed to get value of counter metric, error: ", err)
			}
			log.Println("Successful counter: ", metric, newValue)
		case repository.TypeGauge:
			value, err := strconv.ParseFloat(chi.URLParam(request, "value"), 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err = storage.ReplaceGaugeMetric(metric, value); err != nil {
				log.Println("Failed to change value of gauge metric, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newValue, err := storage.GetValue(models.Gauge, metric)
			if err != nil {
				log.Println("Failed to get value of gauge metric, error: ", err)
			}
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

func ViewMetricValue(storage repository.StorageRepositoryI) http.HandlerFunc {
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
