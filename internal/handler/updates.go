package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"github.com/Bessima/metrics-collect/internal/cryptomessage"
	"io"
	"net/http"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
)

// UpdatesHandler обновляет или сохраняет метрики, переданные через json-параметры
type UpdatesHandler struct {
	storage         repository.StorageRepositorier
	metricsFromFile *repository.MetricsFromFile
	auditEvent      *audit.Event
	privateKey      *rsa.PrivateKey
}

func NewUpdatesHandler(
	storage repository.StorageRepositorier,
	metricsFromFile *repository.MetricsFromFile,
	auditEvent *audit.Event,
	privateKey *rsa.PrivateKey,
) *UpdatesHandler {
	return &UpdatesHandler{
		storage:         storage,
		metricsFromFile: metricsFromFile,
		auditEvent:      auditEvent,
		privateKey:      privateKey,
	}
}

func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics
	var buf bytes.Buffer
	var metricsNames []string

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data []byte

	if h.privateKey != nil {
		data, err = cryptomessage.DecryptMessage(buf.Bytes(), h.privateKey)
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
		err := updateMetricInStorage(h.storage, metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	if h.metricsFromFile != nil {
		repository.UpdateMetricInFile(h.storage, h.metricsFromFile)
	}

	if h.auditEvent != nil {
		h.auditEvent.Notify(metricsNames, r.RemoteAddr)
	}
}

func (h *UpdatesHandler) Handler() http.HandlerFunc {
	return h.ServeHTTP
}
