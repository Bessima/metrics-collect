package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock storage for testing UpdateMetricInFile
type mockStorage struct {
	metrics []models.Metrics
	allErr  error
}

func (m *mockStorage) Counter(name string, value int64) error {
	return nil
}

func (m *mockStorage) ReplaceGaugeMetric(name string, value float64) error {
	return nil
}

func (m *mockStorage) GetValue(typeMetric TypeMetric, name string) (interface{}, error) {
	return nil, nil
}

func (m *mockStorage) GetMetric(typeMetric TypeMetric, name string) (models.Metrics, error) {
	return models.Metrics{}, nil
}

func (m *mockStorage) Load(metrics []models.Metrics) error {
	return nil
}

func (m *mockStorage) All() ([]models.Metrics, error) {
	if m.allErr != nil {
		return nil, m.allErr
	}
	return m.metrics, nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) Ping(ctx context.Context) error {
	return nil
}

func TestUpdateMetricInFile_NilMetricsFromFile(t *testing.T) {
	storage := &mockStorage{
		metrics: []models.Metrics{},
	}

	// Should not panic with nil metricsFromFile
	UpdateMetricInFile(storage, nil)

	// No assertions needed, just checking it doesn't panic
}

func TestUpdateMetricInFile_Success(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "update_test.json")

	metricsFromFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFromFile)

	counterValue := int64(50)
	gaugeValue := 7.89

	storage := &mockStorage{
		metrics: []models.Metrics{
			{
				ID:    "counter1",
				MType: models.Counter,
				Delta: &counterValue,
			},
			{
				ID:    "gauge1",
				MType: models.Gauge,
				Value: &gaugeValue,
			},
		},
	}

	UpdateMetricInFile(storage, metricsFromFile)

	// Verify metrics were written to file
	loadedMetricsFile := &MetricsFromFile{FileName: filename}
	err := loadedMetricsFile.Load()
	require.NoError(t, err)

	loadedMetrics := loadedMetricsFile.GetMetrics()
	assert.Equal(t, 2, len(loadedMetrics))
}

func TestUpdateMetricInFile_StorageAllError(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "error_test.json")

	metricsFromFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFromFile)

	storage := &mockStorage{
		allErr: errors.New("storage error"),
	}

	// Should not panic even with error
	UpdateMetricInFile(storage, metricsFromFile)

	// No metrics should be written
	loadedMetricsFile := &MetricsFromFile{FileName: filename}
	err := loadedMetricsFile.Load()
	// File should be empty or have error
	assert.Error(t, err)
}

func TestUpdateMetricInFile_EmptyMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "empty_test.json")

	metricsFromFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFromFile)

	storage := &mockStorage{
		metrics: []models.Metrics{},
	}

	UpdateMetricInFile(storage, metricsFromFile)

	// Verify empty array was written
	loadedMetricsFile := &MetricsFromFile{FileName: filename}
	err := loadedMetricsFile.Load()
	require.NoError(t, err)

	loadedMetrics := loadedMetricsFile.GetMetrics()
	assert.Equal(t, 0, len(loadedMetrics))
}

func TestUpdateMetricInFile_MultipleMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "multiple_test.json")

	metricsFromFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFromFile)

	counter1 := int64(10)
	counter2 := int64(20)
	counter3 := int64(30)
	gauge1 := 1.1
	gauge2 := 2.2
	gauge3 := 3.3

	storage := &mockStorage{
		metrics: []models.Metrics{
			{ID: "counter1", MType: models.Counter, Delta: &counter1},
			{ID: "counter2", MType: models.Counter, Delta: &counter2},
			{ID: "counter3", MType: models.Counter, Delta: &counter3},
			{ID: "gauge1", MType: models.Gauge, Value: &gauge1},
			{ID: "gauge2", MType: models.Gauge, Value: &gauge2},
			{ID: "gauge3", MType: models.Gauge, Value: &gauge3},
		},
	}

	UpdateMetricInFile(storage, metricsFromFile)

	// Verify all metrics were written
	loadedMetricsFile := &MetricsFromFile{FileName: filename}
	err := loadedMetricsFile.Load()
	require.NoError(t, err)

	loadedMetrics := loadedMetricsFile.GetMetrics()
	assert.Equal(t, 6, len(loadedMetrics))

	// Verify each metric
	metricMap := make(map[string]models.Metrics)
	for _, m := range loadedMetrics {
		metricMap[m.ID] = m
	}

	assert.Equal(t, int64(10), *metricMap["counter1"].Delta)
	assert.Equal(t, int64(20), *metricMap["counter2"].Delta)
	assert.Equal(t, int64(30), *metricMap["counter3"].Delta)
	assert.Equal(t, 1.1, *metricMap["gauge1"].Value)
	assert.Equal(t, 2.2, *metricMap["gauge2"].Value)
	assert.Equal(t, 3.3, *metricMap["gauge3"].Value)
}
