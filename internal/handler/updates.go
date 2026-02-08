package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
)

// UpdatesHandler обновляет или сохраняет метрики, переданные через json-параметры
func UpdatesHandler(storage repository.StorageRepositorier, metricsFromFile *repository.MetricsFromFile, auditEvent *audit.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		var buf bytes.Buffer
		var metricsNames []string

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
			metricsNames = append(metricsNames, metric.ID)
			err := updateMetricInStorage(storage, metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}

		if metricsFromFile != nil {
			repository.UpdateMetricInFile(storage, metricsFromFile)
		}

		if auditEvent != nil {
			auditEvent.Notify(metricsNames, r.RemoteAddr)
		}
	}
}
