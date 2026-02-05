package repository

import (
	"context"
	"testing"
	"time"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.counters)
	assert.NotNil(t, storage.gauge)
	assert.Equal(t, 0, len(storage.counters))
	assert.Equal(t, 0, len(storage.gauge))
}

func TestMemStorage_Counter(t *testing.T) {
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
			addValue:      10,
			expectedValue: 10,
			metricName:    "test_counter",
		},
		{
			name:          "increment existing counter",
			initialValue:  ptr(int64(5)),
			addValue:      3,
			expectedValue: 8,
			metricName:    "existing_counter",
		},
		{
			name:          "add negative value",
			initialValue:  ptr(int64(10)),
			addValue:      -3,
			expectedValue: 7,
			metricName:    "negative_counter",
		},
		{
			name:          "add zero value",
			initialValue:  ptr(int64(5)),
			addValue:      0,
			expectedValue: 5,
			metricName:    "zero_counter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			// Setup initial value if provided
			if tt.initialValue != nil {
				storage.counters[tt.metricName] = models.Metrics{
					ID:    tt.metricName,
					MType: models.Counter,
					Delta: tt.initialValue,
				}
			}

			err := storage.Counter(tt.metricName, tt.addValue)
			require.NoError(t, err)

			value, err := storage.GetValue(TypeCounter, tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestMemStorage_ReplaceGaugeMetric(t *testing.T) {
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
			newValue:      3.14,
			expectedValue: 3.14,
			metricName:    "test_gauge",
		},
		{
			name:          "replace existing gauge",
			initialValue:  ptr(2.5),
			newValue:      5.7,
			expectedValue: 5.7,
			metricName:    "existing_gauge",
		},
		{
			name:          "replace with zero",
			initialValue:  ptr(10.5),
			newValue:      0.0,
			expectedValue: 0.0,
			metricName:    "zero_gauge",
		},
		{
			name:          "replace with negative value",
			initialValue:  ptr(5.0),
			newValue:      -3.14,
			expectedValue: -3.14,
			metricName:    "negative_gauge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			// Setup initial value if provided
			if tt.initialValue != nil {
				storage.gauge[tt.metricName] = models.Metrics{
					ID:    tt.metricName,
					MType: models.Gauge,
					Value: tt.initialValue,
				}
			}

			err := storage.ReplaceGaugeMetric(tt.metricName, tt.newValue)
			require.NoError(t, err)

			value, err := storage.GetValue(TypeGauge, tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestMemStorage_GetValue(t *testing.T) {
	storage := NewMemStorage()

	// Setup test data
	counterValue := int64(42)
	gaugeValue := 3.14

	storage.counters["test_counter"] = models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: &counterValue,
	}

	storage.gauge["test_gauge"] = models.Metrics{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	tests := []struct {
		name        string
		metricType  TypeMetric
		metricName  string
		expectValue interface{}
		expectError bool
	}{
		{
			name:        "get existing counter",
			metricType:  TypeCounter,
			metricName:  "test_counter",
			expectValue: int64(42),
			expectError: false,
		},
		{
			name:        "get existing gauge",
			metricType:  TypeGauge,
			metricName:  "test_gauge",
			expectValue: 3.14,
			expectError: false,
		},
		{
			name:        "get non-existing counter",
			metricType:  TypeCounter,
			metricName:  "non_existing",
			expectValue: nil,
			expectError: true,
		},
		{
			name:        "get non-existing gauge",
			metricType:  TypeGauge,
			metricName:  "non_existing",
			expectValue: nil,
			expectError: true,
		},
		{
			name:        "get with unknown metric type",
			metricType:  TypeMetric("unknown"),
			metricName:  "test",
			expectValue: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := storage.GetValue(tt.metricType, tt.metricName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, value)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	storage := NewMemStorage()

	// Setup test data
	counterValue := int64(42)
	gaugeValue := 3.14

	storage.counters["test_counter"] = models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: &counterValue,
	}

	storage.gauge["test_gauge"] = models.Metrics{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	tests := []struct {
		name         string
		metricType   TypeMetric
		metricName   string
		expectMetric models.Metrics
		expectError  bool
	}{
		{
			name:       "get existing counter metric",
			metricType: TypeCounter,
			metricName: "test_counter",
			expectMetric: models.Metrics{
				ID:    "test_counter",
				MType: models.Counter,
				Delta: &counterValue,
			},
			expectError: false,
		},
		{
			name:       "get existing gauge metric",
			metricType: TypeGauge,
			metricName: "test_gauge",
			expectMetric: models.Metrics{
				ID:    "test_gauge",
				MType: models.Gauge,
				Value: &gaugeValue,
			},
			expectError: false,
		},
		{
			name:         "get non-existing counter",
			metricType:   TypeCounter,
			metricName:   "non_existing",
			expectMetric: models.Metrics{},
			expectError:  true,
		},
		{
			name:         "get with unknown type",
			metricType:   TypeMetric("unknown"),
			metricName:   "test",
			expectMetric: models.Metrics{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := storage.GetMetric(tt.metricType, tt.metricName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectMetric.ID, metric.ID)
				assert.Equal(t, tt.expectMetric.MType, metric.MType)

				if tt.expectMetric.Delta != nil {
					require.NotNil(t, metric.Delta)
					assert.Equal(t, *tt.expectMetric.Delta, *metric.Delta)
				}

				if tt.expectMetric.Value != nil {
					require.NotNil(t, metric.Value)
					assert.Equal(t, *tt.expectMetric.Value, *metric.Value)
				}
			}
		})
	}
}

func TestMemStorage_All(t *testing.T) {
	storage := NewMemStorage()

	// Empty storage
	metrics, err := storage.All()
	require.NoError(t, err)
	assert.Equal(t, 0, len(metrics))

	// Add some metrics
	counterValue1 := int64(10)
	counterValue2 := int64(20)
	gaugeValue1 := 3.14
	gaugeValue2 := 2.71

	storage.counters["counter1"] = models.Metrics{
		ID:    "counter1",
		MType: models.Counter,
		Delta: &counterValue1,
	}
	storage.counters["counter2"] = models.Metrics{
		ID:    "counter2",
		MType: models.Counter,
		Delta: &counterValue2,
	}
	storage.gauge["gauge1"] = models.Metrics{
		ID:    "gauge1",
		MType: models.Gauge,
		Value: &gaugeValue1,
	}
	storage.gauge["gauge2"] = models.Metrics{
		ID:    "gauge2",
		MType: models.Gauge,
		Value: &gaugeValue2,
	}

	metrics, err = storage.All()
	require.NoError(t, err)
	assert.Equal(t, 4, len(metrics))

	// Check that all metrics are present
	metricNames := make(map[string]bool)
	for _, m := range metrics {
		metricNames[m.ID] = true
	}

	assert.True(t, metricNames["counter1"])
	assert.True(t, metricNames["counter2"])
	assert.True(t, metricNames["gauge1"])
	assert.True(t, metricNames["gauge2"])
}

func TestMemStorage_Load(t *testing.T) {
	storage := NewMemStorage()

	counterValue1 := int64(100)
	counterValue2 := int64(200)
	gaugeValue1 := 1.23
	gaugeValue2 := 4.56

	metricsToLoad := []models.Metrics{
		{
			ID:    "loaded_counter1",
			MType: models.Counter,
			Delta: &counterValue1,
		},
		{
			ID:    "loaded_counter2",
			MType: models.Counter,
			Delta: &counterValue2,
		},
		{
			ID:    "loaded_gauge1",
			MType: models.Gauge,
			Value: &gaugeValue1,
		},
		{
			ID:    "loaded_gauge2",
			MType: models.Gauge,
			Value: &gaugeValue2,
		},
	}

	err := storage.Load(metricsToLoad)
	require.NoError(t, err)

	// Verify counters
	assert.Equal(t, 2, len(storage.counters))
	counter1, exists := storage.counters["loaded_counter1"]
	assert.True(t, exists)
	assert.Equal(t, int64(100), *counter1.Delta)

	counter2, exists := storage.counters["loaded_counter2"]
	assert.True(t, exists)
	assert.Equal(t, int64(200), *counter2.Delta)

	// Verify gauges
	assert.Equal(t, 2, len(storage.gauge))
	gauge1, exists := storage.gauge["loaded_gauge1"]
	assert.True(t, exists)
	assert.Equal(t, 1.23, *gauge1.Value)

	gauge2, exists := storage.gauge["loaded_gauge2"]
	assert.True(t, exists)
	assert.Equal(t, 4.56, *gauge2.Value)
}

func TestMemStorage_Ping(t *testing.T) {
	storage := NewMemStorage()

	err := storage.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only for DB")
}

func TestMemStorage_Close(t *testing.T) {
	storage := NewMemStorage()

	err := storage.Close()
	assert.NoError(t, err)
}

// TestMemStorage_ConcurrentCounterCreation проверяет race condition при создании новых счетчиков
// Запустите с флагом -race для обнаружения проблемы:
//
//	go test -race -run=TestMemStorage_ConcurrentCounterCreation ./internal/repository/
func TestMemStorage_ConcurrentCounterCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	storage := NewMemStorage()
	numGoroutines := 100
	numOperations := 1000

	// Создаем канал для синхронизации старта всех горутин
	start := make(chan struct{})
	done := make(chan struct{})

	// Запускаем горутины, которые будут создавать НОВЫЕ метрики
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			<-start // Ждем сигнала для старта
			for j := 0; j < numOperations; j++ {
				// Каждая горутина создает свои уникальные метрики
				metricName := "counter_goroutine_" + string(rune(id)) + "_" + string(rune(j))
				_ = storage.Counter(metricName, 1)
			}
			done <- struct{}{}
		}(i)
	}

	// Даем сигнал всем горутинам начать одновременно
	close(start)

	// Ждем завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Проверяем, что все метрики были созданы
	metrics, err := storage.All()
	require.NoError(t, err)
	expectedCount := numGoroutines * numOperations
	assert.Equal(t, expectedCount, len(metrics), "Expected %d metrics, got %d", expectedCount, len(metrics))
}

// TestMemStorage_ConcurrentAllRead проверяет race condition в методе All()
// при одновременном чтении и записи
func TestMemStorage_ConcurrentAllRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	storage := NewMemStorage()
	stopChan := make(chan struct{})

	// Предварительно создаем метрики
	for i := 0; i < 100; i++ {
		_ = storage.Counter("counter_"+string(rune(i)), int64(i))
	}

	// Горутина, которая постоянно читает все метрики
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				_, _ = storage.All()
			}
		}
	}()

	// Горутина, которая постоянно обновляет метрики
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				_ = storage.Counter("counter_50", 1)
			}
		}
	}()

	// Запускаем на некоторое время
	time.Sleep(100 * time.Millisecond)
	close(stopChan)

	// Даем время горутинам завершиться
	time.Sleep(10 * time.Millisecond)
}

// Helper function to create pointers
func ptr[T any](v T) *T {
	return &v
}

// ==================== BENCHMARKS ====================

// BenchmarkMemStorageCounter_Sequential измеряет производительность последовательных операций Counter
func BenchmarkMemStorageCounter_Sequential(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.Counter("test_counter", 1)
	}
}

// BenchmarkMemStorageCounter_NewMetrics измеряет создание новых метрик (worst case - без блокировок)
func BenchmarkMemStorageCounter_NewMetrics(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metricName := "counter_" + string(rune(i))
		_ = storage.Counter(metricName, 1)
	}
}

// BenchmarkMemStorageCounter_Parallel измеряет конкурентный доступ к одной метрике
func BenchmarkMemStorageCounter_Parallel(b *testing.B) {
	storage := NewMemStorage()

	// Предварительно создаем метрику
	_ = storage.Counter("shared_counter", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = storage.Counter("shared_counter", 1)
		}
	})
}

// BenchmarkMemStorageCounter_ParallelMultiple измеряет конкурентный доступ к разным метрикам
func BenchmarkMemStorageCounter_ParallelMultiple(b *testing.B) {
	storage := NewMemStorage()
	numMetrics := 100

	// Предварительно создаем метрики
	for i := 0; i < numMetrics; i++ {
		_ = storage.Counter("counter_"+string(rune(i)), 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			metricName := "counter_" + string(rune(i%numMetrics))
			_ = storage.Counter(metricName, 1)
			i++
		}
	})
}

// BenchmarkMemStorageGauge_Sequential измеряет производительность последовательных операций Gauge
func BenchmarkMemStorageGauge_Sequential(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.ReplaceGaugeMetric("test_gauge", 3.14)
	}
}

// BenchmarkMemStorageGauge_Parallel измеряет конкурентный доступ к одной gauge-метрике
func BenchmarkMemStorageGauge_Parallel(b *testing.B) {
	storage := NewMemStorage()

	// Предварительно создаем метрику
	_ = storage.ReplaceGaugeMetric("shared_gauge", 0.0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = storage.ReplaceGaugeMetric("shared_gauge", 3.14)
		}
	})
}

// BenchmarkMemStorageGetValue измеряет производительность чтения метрик
func BenchmarkMemStorageGetValue(b *testing.B) {
	storage := NewMemStorage()
	_ = storage.Counter("test_counter", 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetValue(TypeCounter, "test_counter")
	}
}

// BenchmarkMemStorageGetValue_Parallel измеряет конкурентное чтение
func BenchmarkMemStorageGetValue_Parallel(b *testing.B) {
	storage := NewMemStorage()
	_ = storage.Counter("test_counter", 42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = storage.GetValue(TypeCounter, "test_counter")
		}
	})
}

// BenchmarkMemStorageAll измеряет производительность получения всех метрик
func BenchmarkMemStorageAll(b *testing.B) {
	benchmarks := []struct {
		name        string
		numCounters int
		numGauges   int
	}{
		{"10_metrics", 5, 5},
		{"100_metrics", 50, 50},
		{"1000_metrics", 500, 500},
		{"10000_metrics", 5000, 5000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := NewMemStorage()

			// Заполняем хранилище
			for i := 0; i < bm.numCounters; i++ {
				_ = storage.Counter("counter_"+string(rune(i)), int64(i))
			}
			for i := 0; i < bm.numGauges; i++ {
				_ = storage.ReplaceGaugeMetric("gauge_"+string(rune(i)), float64(i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = storage.All()
			}
		})
	}
}

// BenchmarkMemStorageLoad измеряет производительность загрузки метрик
func BenchmarkMemStorageLoad(b *testing.B) {
	benchmarks := []struct {
		name       string
		numMetrics int
	}{
		{"10_metrics", 10},
		{"100_metrics", 100},
		{"1000_metrics", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Подготавливаем данные для загрузки
			metrics := make([]models.Metrics, bm.numMetrics)
			for i := 0; i < bm.numMetrics/2; i++ {
				val := int64(i)
				metrics[i] = models.Metrics{
					ID:    "counter_" + string(rune(i)),
					MType: models.Counter,
					Delta: &val,
				}
			}
			for i := bm.numMetrics / 2; i < bm.numMetrics; i++ {
				val := float64(i)
				metrics[i] = models.Metrics{
					ID:    "gauge_" + string(rune(i)),
					MType: models.Gauge,
					Value: &val,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				storage := NewMemStorage()
				_ = storage.Load(metrics)
			}
		})
	}
}

// BenchmarkMemStorageMixed_Sequential измеряет смешанные операции (запись/чтение)
func BenchmarkMemStorageMixed_Sequential(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 50% записи, 50% чтения
		if i%2 == 0 {
			_ = storage.Counter("counter", 1)
		} else {
			_, _ = storage.GetValue(TypeCounter, "counter")
		}
	}
}

// BenchmarkMemStorageMixed_Parallel измеряет конкурентные смешанные операции
func BenchmarkMemStorageMixed_Parallel(b *testing.B) {
	storage := NewMemStorage()
	_ = storage.Counter("counter", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 50% записи, 50% чтения
			if i%2 == 0 {
				_ = storage.Counter("counter", 1)
			} else {
				_, _ = storage.GetValue(TypeCounter, "counter")
			}
			i++
		}
	})
}

// BenchmarkMemStorageContentionHigh измеряет производительность при высокой конкуренции
// (много горутин обновляют одну и ту же метрику)
func BenchmarkMemStorageContentionHigh(b *testing.B) {
	storage := NewMemStorage()
	_ = storage.Counter("hotspot", 0)

	b.ResetTimer()
	b.SetParallelism(100) // Высокий уровень параллелизма
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = storage.Counter("hotspot", 1)
		}
	})
}

// BenchmarkMemStorageContentionLow измеряет производительность при низкой конкуренции
// (каждая горутина работает со своей метрикой)
func BenchmarkMemStorageContentionLow(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		goroutineID := 0
		for pb.Next() {
			metricName := "counter_goroutine_" + string(rune(goroutineID))
			_ = storage.Counter(metricName, 1)
		}
	})
}
