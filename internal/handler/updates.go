package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"github.com/Bessima/metrics-collect/internal/crypto_message"
	"io"
	"net/http"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
)

// UpdatesHandler обновляет или сохраняет метрики, переданные через json-параметры
func UpdatesHandler(storage repository.StorageRepositorier, metricsFromFile *repository.MetricsFromFile, auditEvent *audit.Event, privateKey *rsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		var buf bytes.Buffer
		var metricsNames []string

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var data []byte

		if privateKey != nil {
			data, err = crypto_message.DecryptMessage(buf.Bytes(), privateKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			gr, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err = io.ReadAll(gr)
			gr.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if data == nil {
			data = buf.Bytes()
		}

		if err = json.Unmarshal(data, &metrics); err != nil {
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
