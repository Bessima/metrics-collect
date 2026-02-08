package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
)

// UpdateHandler обновляет или сохраняет одну метрику, переданную через json-параметры
func UpdateHandler(storage repository.StorageRepositorier, metricsFromFile *repository.MetricsFromFile) http.HandlerFunc {
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

		err = updateMetricInStorage(storage, metric)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if metricsFromFile != nil {
			repository.UpdateMetricInFile(storage, metricsFromFile)
		}
	}
}
