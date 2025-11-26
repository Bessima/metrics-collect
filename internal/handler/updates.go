package handler

import (
	"bytes"
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"net/http"
)

func UpdatesHandler(storage repository.StorageRepositoryI, metricsFromFile *repository.MetricsFromFile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, metric := range metrics {
			err := updateMetricInStorage(storage, metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}

		if metricsFromFile != nil {
			repository.UpdateMetricInFile(storage, metricsFromFile)
		}
	}
}
