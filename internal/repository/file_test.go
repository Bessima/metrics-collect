package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsFromFile(t *testing.T) {
	t.Run("create with valid filename", func(t *testing.T) {
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "test_metrics.json")

		metricsFile := NewMetricsFromFile(filename)

		assert.NotNil(t, metricsFile)
		assert.Equal(t, filename, metricsFile.FileName)

		// Check file was created
		_, err := os.Stat(filename)
		assert.NoError(t, err)
	})

	t.Run("create with empty filename", func(t *testing.T) {
		metricsFile := NewMetricsFromFile("")

		assert.Nil(t, metricsFile)
	})
}

func TestMetricsFromFile_UpdateMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "metrics.json")

	metricsFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFile)

	counterValue := int64(42)
	gaugeValue := 3.14

	testMetrics := []models.Metrics{
		{
			ID:    "test_counter",
			MType: models.Counter,
			Delta: &counterValue,
		},
		{
			ID:    "test_gauge",
			MType: models.Gauge,
			Value: &gaugeValue,
		},
	}

	err := metricsFile.UpdateMetrics(&testMetrics)
	require.NoError(t, err)

	// Verify file exists and has content
	fileData, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Greater(t, len(fileData), 0)
}

func TestMetricsFromFile_Load(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "metrics.json")

	// Create test data
	counterValue := int64(100)
	gaugeValue := 5.67

	initialMetrics := []models.Metrics{
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
	}

	// Write test data
	metricsFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFile)
	err := metricsFile.UpdateMetrics(&initialMetrics)
	require.NoError(t, err)

	// Create new instance and load
	newMetricsFile := &MetricsFromFile{FileName: filename}
	err = newMetricsFile.Load()
	require.NoError(t, err)

	loadedMetrics := newMetricsFile.GetMetrics()
	assert.Equal(t, 2, len(loadedMetrics))

	// Verify loaded data
	for _, m := range loadedMetrics {
		if m.ID == "counter1" {
			assert.Equal(t, models.Counter, m.MType)
			require.NotNil(t, m.Delta)
			assert.Equal(t, int64(100), *m.Delta)
		} else if m.ID == "gauge1" {
			assert.Equal(t, models.Gauge, m.MType)
			require.NotNil(t, m.Value)
			assert.Equal(t, 5.67, *m.Value)
		}
	}
}

func TestMetricsFromFile_Load_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "empty_metrics.json")

	// Create empty file
	err := os.WriteFile(filename, []byte(""), 0666)
	require.NoError(t, err)

	metricsFile := &MetricsFromFile{FileName: filename}
	err = metricsFile.Load()

	assert.Error(t, err)
}

func TestMetricsFromFile_Load_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "invalid_metrics.json")

	// Create file with invalid JSON
	err := os.WriteFile(filename, []byte("not a valid json"), 0666)
	require.NoError(t, err)

	metricsFile := &MetricsFromFile{FileName: filename}
	err = metricsFile.Load()

	assert.Error(t, err)
}

func TestMetricsFromFile_GetMetrics(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "metrics.json")

	metricsFile := NewMetricsFromFile(filename)
	require.NotNil(t, metricsFile)

	counterValue := int64(50)
	testMetrics := []models.Metrics{
		{
			ID:    "test",
			MType: models.Counter,
			Delta: &counterValue,
		},
	}

	metricsFile.metrics = testMetrics

	result := metricsFile.GetMetrics()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test", result[0].ID)
}

func TestNewFileStorageRepository(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "storage.json")

	repo := NewFileStorageRepository(filename)

	assert.NotNil(t, repo)
	assert.Equal(t, filename, repo.FileName)

	// Check file was created
	_, err := os.Stat(filename)
	assert.NoError(t, err)
}

func TestFileStorageRepository_Counter(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  *int64
		addValue      int64
		expectedValue int64
		metricName    string
	}{
		{
			name:          "add counter to empty storage",
			initialValue:  nil,
			addValue:      15,
			expectedValue: 15,
			metricName:    "counter1",
		},
		{
			name:          "increment existing counter",
			initialValue:  ptr(int64(10)),
			addValue:      5,
			expectedValue: 15,
			metricName:    "counter2",
		},
		{
			name:          "add negative value",
			initialValue:  ptr(int64(20)),
			addValue:      -5,
			expectedValue: 15,
			metricName:    "counter3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, "counter_test.json")

			repo := NewFileStorageRepository(filename)

			// Setup initial value if provided
			if tt.initialValue != nil {
				initialMetrics := []models.Metrics{
					{
						ID:    tt.metricName,
						MType: string(TypeCounter),
						Delta: tt.initialValue,
					},
				}
				err := repo.Load(initialMetrics)
				require.NoError(t, err)
			}

			err := repo.Counter(tt.metricName, tt.addValue)
			require.NoError(t, err)

			value, err := repo.GetValue(TypeCounter, tt.metricName)
			require.NoError(t, err)

			deltaValue, ok := value.(*int64)
			require.True(t, ok)
			assert.Equal(t, tt.expectedValue, *deltaValue)
		})
	}
}

func TestFileStorageRepository_ReplaceGaugeMetric(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  *float64
		newValue      float64
		expectedValue float64
		metricName    string
	}{
		{
			name:          "add gauge to empty storage",
			initialValue:  nil,
			newValue:      2.5,
			expectedValue: 2.5,
			metricName:    "gauge1",
		},
		{
			name:          "replace existing gauge",
			initialValue:  ptr(1.5),
			newValue:      3.7,
			expectedValue: 3.7,
			metricName:    "gauge2",
		},
		{
			name:          "replace with zero",
			initialValue:  ptr(10.0),
			newValue:      0.0,
			expectedValue: 0.0,
			metricName:    "gauge3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, "gauge_test.json")

			repo := NewFileStorageRepository(filename)

			// Setup initial value if provided
			if tt.initialValue != nil {
				initialMetrics := []models.Metrics{
					{
						ID:    tt.metricName,
						MType: string(TypeGauge),
						Value: tt.initialValue,
					},
				}
				err := repo.Load(initialMetrics)
				require.NoError(t, err)
			}

			err := repo.ReplaceGaugeMetric(tt.metricName, tt.newValue)
			require.NoError(t, err)

			value, err := repo.GetValue(TypeGauge, tt.metricName)
			require.NoError(t, err)

			gaugeValue, ok := value.(*float64)
			require.True(t, ok)
			assert.Equal(t, tt.expectedValue, *gaugeValue)
		})
	}
}

func TestFileStorageRepository_GetValue(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "get_value_test.json")

	repo := NewFileStorageRepository(filename)

	// Setup test data
	counterValue := int64(42)
	gaugeValue := 3.14

	initialMetrics := []models.Metrics{
		{
			ID:    "test_counter",
			MType: string(TypeCounter),
			Delta: &counterValue,
		},
		{
			ID:    "test_gauge",
			MType: string(TypeGauge),
			Value: &gaugeValue,
		},
	}

	err := repo.Load(initialMetrics)
	require.NoError(t, err)

	tests := []struct {
		name        string
		metricType  TypeMetric
		metricName  string
		expectError bool
	}{
		{
			name:        "get existing counter",
			metricType:  TypeCounter,
			metricName:  "test_counter",
			expectError: false,
		},
		{
			name:        "get existing gauge",
			metricType:  TypeGauge,
			metricName:  "test_gauge",
			expectError: false,
		},
		{
			name:        "get non-existing counter",
			metricType:  TypeCounter,
			metricName:  "non_existing",
			expectError: true,
		},
		{
			name:        "get with unknown type",
			metricType:  TypeMetric("unknown"),
			metricName:  "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := repo.GetValue(tt.metricType, tt.metricName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, value)
			}
		})
	}
}

func TestFileStorageRepository_GetMetric(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "get_metric_test.json")

	repo := NewFileStorageRepository(filename)

	counterValue := int64(100)
	gaugeValue := 2.71

	initialMetrics := []models.Metrics{
		{
			ID:    "counter_metric",
			MType: string(TypeCounter),
			Delta: &counterValue,
		},
		{
			ID:    "gauge_metric",
			MType: string(TypeGauge),
			Value: &gaugeValue,
		},
	}

	err := repo.Load(initialMetrics)
	require.NoError(t, err)

	t.Run("get existing counter metric", func(t *testing.T) {
		metric, err := repo.GetMetric(TypeCounter, "counter_metric")
		require.NoError(t, err)
		assert.Equal(t, "counter_metric", metric.ID)
		assert.Equal(t, string(TypeCounter), metric.MType)
		require.NotNil(t, metric.Delta)
		assert.Equal(t, int64(100), *metric.Delta)
	})

	t.Run("get existing gauge metric", func(t *testing.T) {
		metric, err := repo.GetMetric(TypeGauge, "gauge_metric")
		require.NoError(t, err)
		assert.Equal(t, "gauge_metric", metric.ID)
		assert.Equal(t, string(TypeGauge), metric.MType)
		require.NotNil(t, metric.Value)
		assert.Equal(t, 2.71, *metric.Value)
	})

	t.Run("get non-existing metric", func(t *testing.T) {
		_, err := repo.GetMetric(TypeCounter, "non_existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestFileStorageRepository_All(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "all_test.json")

	repo := NewFileStorageRepository(filename)

	t.Run("empty storage", func(t *testing.T) {
		metrics, err := repo.All()
		require.NoError(t, err)
		assert.Equal(t, 0, len(metrics))
	})

	t.Run("storage with metrics", func(t *testing.T) {
		counterValue1 := int64(10)
		counterValue2 := int64(20)
		gaugeValue1 := 1.1
		gaugeValue2 := 2.2

		initialMetrics := []models.Metrics{
			{
				ID:    "counter1",
				MType: string(TypeCounter),
				Delta: &counterValue1,
			},
			{
				ID:    "counter2",
				MType: string(TypeCounter),
				Delta: &counterValue2,
			},
			{
				ID:    "gauge1",
				MType: string(TypeGauge),
				Value: &gaugeValue1,
			},
			{
				ID:    "gauge2",
				MType: string(TypeGauge),
				Value: &gaugeValue2,
			},
		}

		err := repo.Load(initialMetrics)
		require.NoError(t, err)

		metrics, err := repo.All()
		require.NoError(t, err)
		assert.Equal(t, 4, len(metrics))
	})
}

func TestFileStorageRepository_Load(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "load_test.json")

	repo := NewFileStorageRepository(filename)

	counterValue := int64(99)
	gaugeValue := 8.88

	metricsToLoad := []models.Metrics{
		{
			ID:    "loaded_counter",
			MType: string(TypeCounter),
			Delta: &counterValue,
		},
		{
			ID:    "loaded_gauge",
			MType: string(TypeGauge),
			Value: &gaugeValue,
		},
	}

	err := repo.Load(metricsToLoad)
	require.NoError(t, err)

	// Verify metrics were saved
	loadedMetrics, err := repo.All()
	require.NoError(t, err)
	assert.Equal(t, 2, len(loadedMetrics))
}

func TestFileStorageRepository_Ping(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "ping_test.json")

	repo := NewFileStorageRepository(filename)

	err := repo.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only for DB")
}

func TestFileStorageRepository_Close(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "close_test.json")

	repo := NewFileStorageRepository(filename)

	err := repo.Close()
	assert.NoError(t, err)
}
