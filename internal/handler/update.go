package handler

import (
	"bytes"
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
	"strconv"
)

func UpdateHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.ShortFieldsMetric
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
			value, _ := strconv.ParseInt(metric.Value, 10, 64)
			storage.Counter(metric.ID, value)
			newValue, _ := storage.GetValue(models.Counter, metric.ID)
			log.Println("Successful counter: ", metric.ID, newValue)
		case repository.TypeGauge:
			value, _ := strconv.ParseFloat(metric.Value, 64)
			storage.ReplaceGaugeMetric(metric.ID, value)
			newValue, _ := storage.GetValue(models.Gauge, metric.ID)
			log.Println("Successful replacing gauge: ", metric.ID, newValue)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	}
}
