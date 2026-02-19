package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatesHandler_Success(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	counter1 := int64(10)
	counter2 := int64(20)
	gauge1 := 3.14
	gauge2 := 2.71

	metrics := []models.Metrics{
		{
			ID:    "counter1",
			MType: models.Counter,
			Delta: &counter1,
		},
		{
			ID:    "counter2",
			MType: models.Counter,
			Delta: &counter2,
		},
		{
			ID:    "gauge1",
			MType: models.Gauge,
			Value: &gauge1,
		},
		{
			ID:    "gauge2",
			MType: models.Gauge,
			Value: &gauge2,
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify metrics were saved
	val, err := storage.GetValue(repository.TypeCounter, "counter1")
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)

	val, err = storage.GetValue(repository.TypeCounter, "counter2")
	require.NoError(t, err)
	assert.Equal(t, int64(20), val)

	val, err = storage.GetValue(repository.TypeGauge, "gauge1")
	require.NoError(t, err)
	assert.Equal(t, 3.14, val)

	val, err = storage.GetValue(repository.TypeGauge, "gauge2")
	require.NoError(t, err)
	assert.Equal(t, 2.71, val)
}

func TestUpdatesHandler_EmptyArray(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	metrics := []models.Metrics{}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdatesHandler_InvalidJSON(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid")
}

func TestUpdatesHandler_CounterWithoutDelta(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	metrics := []models.Metrics{
		{
			ID:    "counter1",
			MType: models.Counter,
			Delta: nil, // Missing delta
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "delta value not found")
}

func TestUpdatesHandler_GaugeWithoutValue(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	metrics := []models.Metrics{
		{
			ID:    "gauge1",
			MType: models.Gauge,
			Value: nil, // Missing value
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "value not found")
}

func TestUpdatesHandler_UnsupportedType(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	value := int64(10)
	metrics := []models.Metrics{
		{
			ID:    "metric1",
			MType: "unsupported", // Unsupported type
			Delta: &value,
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "not supported")
}

func TestUpdatesHandler_WithMetricsFromFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "metrics.json")

	storage := repository.NewMemStorage()
	metricsFromFile := repository.NewMetricsFromFile(filename)
	handler := UpdatesHandler(storage, metricsFromFile, nil)

	counter := int64(100)
	metrics := []models.Metrics{
		{
			ID:    "counter_file",
			MType: models.Counter,
			Delta: &counter,
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify metric was saved in storage
	val, err := storage.GetValue(repository.TypeCounter, "counter_file")
	require.NoError(t, err)
	assert.Equal(t, int64(100), val)

	// Verify file was updated
	fileMetrics := &repository.MetricsFromFile{FileName: filename}
	err = fileMetrics.Load()
	require.NoError(t, err)

	loadedMetrics := fileMetrics.GetMetrics()
	assert.Equal(t, 1, len(loadedMetrics))
	assert.Equal(t, "counter_file", loadedMetrics[0].ID)
}

func TestUpdatesHandler_WithAuditEvent(t *testing.T) {
	tempDir := t.TempDir()
	auditFile := filepath.Join(tempDir, "audit.log")

	storage := repository.NewMemStorage()
	auditEvent := &audit.Event{}
	fileSubscriber := audit.NewFileSubscriber(auditFile)
	auditEvent.Register(fileSubscriber)

	handler := UpdatesHandler(storage, nil, auditEvent)

	counter := int64(50)
	gauge := 1.23
	metrics := []models.Metrics{
		{
			ID:    "audit_counter",
			MType: models.Counter,
			Delta: &counter,
		},
		{
			ID:    "audit_gauge",
			MType: models.Gauge,
			Value: &gauge,
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify audit log was written
	auditData, err := os.ReadFile(auditFile)
	if err == nil {
		assert.Greater(t, len(auditData), 0)
	}
}

func TestUpdatesHandler_MixedValidAndInvalidMetrics(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	validCounter := int64(10)
	metrics := []models.Metrics{
		{
			ID:    "valid_counter",
			MType: models.Counter,
			Delta: &validCounter,
		},
		{
			ID:    "invalid_counter",
			MType: models.Counter,
			Delta: nil, // Invalid - missing delta
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should return error for invalid metric
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// But first valid metric should be saved
	val, err := storage.GetValue(repository.TypeCounter, "valid_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), val)
}

func TestUpdatesHandler_MultipleCounterIncrements(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// First update
	counter1 := int64(10)
	metrics1 := []models.Metrics{
		{
			ID:    "counter_inc",
			MType: models.Counter,
			Delta: &counter1,
		},
	}

	body1, _ := json.Marshal(metrics1)
	req1 := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body1))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second update
	counter2 := int64(15)
	metrics2 := []models.Metrics{
		{
			ID:    "counter_inc",
			MType: models.Counter,
			Delta: &counter2,
		},
	}

	body2, _ := json.Marshal(metrics2)
	req2 := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body2))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify counter was incremented
	val, err := storage.GetValue(repository.TypeCounter, "counter_inc")
	require.NoError(t, err)
	assert.Equal(t, int64(25), val) // 10 + 15
}

func TestUpdatesHandler_GaugeReplacement(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// First update
	gauge1 := 1.5
	metrics1 := []models.Metrics{
		{
			ID:    "gauge_replace",
			MType: models.Gauge,
			Value: &gauge1,
		},
	}

	body1, _ := json.Marshal(metrics1)
	req1 := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body1))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second update
	gauge2 := 3.7
	metrics2 := []models.Metrics{
		{
			ID:    "gauge_replace",
			MType: models.Gauge,
			Value: &gauge2,
		},
	}

	body2, _ := json.Marshal(metrics2)
	req2 := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body2))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify gauge was replaced (not incremented)
	val, err := storage.GetValue(repository.TypeGauge, "gauge_replace")
	require.NoError(t, err)
	assert.Equal(t, 3.7, val) // Replaced, not 1.5 + 3.7
}

func TestUpdatesHandler_LargePayload(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// Create 100 metrics
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		counter := int64(i)
		metrics[i] = models.Metrics{
			ID:    "counter_" + string(rune(i)),
			MType: models.Counter,
			Delta: &counter,
		}
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdatesHandler_EmptyBody(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdatesHandler_NilBody(t *testing.T) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should handle nil body gracefully
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
