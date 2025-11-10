package handler

import (
	"bytes"
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
)

func UpdateHandler(storage *repository.MemStorage, metricsFromFile *repository.MetricsFromFile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch repository.TypeMetric(metric.MType) {
		case repository.TypeCounter:
			if metric.Delta == nil {
				log.Println("Delta value not found for ", metric.ID)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			delta := *metric.Delta
			storage.Counter(metric.ID, delta)
			newValue, _ := storage.GetValue(models.Counter, metric.ID)
			log.Println("Successful counter: ", metric.ID, newValue)
		case repository.TypeGauge:
			if metric.Value == nil {
				log.Println("Value not found for ", metric.ID)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			value := *metric.Value
			storage.ReplaceGaugeMetric(metric.ID, value)
			newValue, _ := storage.GetValue(models.Gauge, metric.ID)
			log.Println("Successful replacing gauge: ", metric.ID, newValue)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if metricsFromFile != nil {
			repository.UpdateMetricInFile(storage, metricsFromFile)
		}
	}
}
