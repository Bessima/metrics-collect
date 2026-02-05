package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
)

func init() {
	// Отключаем логи для бенчмарков
	log.SetOutput(io.Discard)
}

// ==================== BENCHMARKS ====================

// BenchmarkUpdatesHandler_Sequential измеряет производительность batch-обработки
// с разными размерами payload
func BenchmarkUpdatesHandler_Sequential(b *testing.B) {
	benchmarks := []struct {
		name       string
		numMetrics int
	}{
		{"10_metrics", 10},
		{"50_metrics", 50},
		{"100_metrics", 100},
		{"500_metrics", 500},
		{"1000_metrics", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := repository.NewMemStorage()
			handler := UpdatesHandler(storage, nil, nil)

			// Подготавливаем payload
			metrics := make([]models.Metrics, bm.numMetrics)
			for i := 0; i < bm.numMetrics; i++ {
				if i%2 == 0 {
					// Counter
					delta := int64(i)
					metrics[i] = models.Metrics{
						ID:    "counter_" + strconv.Itoa(i),
						MType: models.Counter,
						Delta: &delta,
					}
				} else {
					// Gauge
					value := float64(i)
					metrics[i] = models.Metrics{
						ID:    "gauge_" + strconv.Itoa(i),
						MType: models.Gauge,
						Value: &value,
					}
				}
			}

			body, err := json.Marshal(metrics)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					b.Fatalf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
				}
			}
		})
	}
}

// BenchmarkUpdatesHandler_JSONUnmarshal измеряет только десериализацию JSON
func BenchmarkUpdatesHandler_JSONUnmarshal(b *testing.B) {
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
			// Подготавливаем JSON payload
			metrics := make([]models.Metrics, bm.numMetrics)
			for i := 0; i < bm.numMetrics; i++ {
				delta := int64(i)
				metrics[i] = models.Metrics{
					ID:    "counter_" + strconv.Itoa(i),
					MType: models.Counter,
					Delta: &delta,
				}
			}

			body, err := json.Marshal(metrics)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			b.SetBytes(int64(len(body)))

			for i := 0; i < b.N; i++ {
				var parsedMetrics []models.Metrics
				err := json.Unmarshal(body, &parsedMetrics)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUpdatesHandler_MetricUpdate измеряет только обновление метрик в storage
// (без HTTP overhead и JSON parsing)
func BenchmarkUpdatesHandler_MetricUpdate(b *testing.B) {
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
			storage := repository.NewMemStorage()

			// Подготавливаем метрики
			metrics := make([]models.Metrics, bm.numMetrics)
			for i := 0; i < bm.numMetrics; i++ {
				delta := int64(i)
				metrics[i] = models.Metrics{
					ID:    "counter_" + strconv.Itoa(i),
					MType: models.Counter,
					Delta: &delta,
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, metric := range metrics {
					err := updateMetricInStorage(storage, metric)
					if err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

// BenchmarkUpdatesHandler_RepeatedUpdates измеряет производительность
// многократных обновлений одних и тех же метрик
func BenchmarkUpdatesHandler_RepeatedUpdates(b *testing.B) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// Создаем 10 метрик
	metrics := make([]models.Metrics, 10)
	for i := 0; i < 10; i++ {
		delta := int64(1)
		metrics[i] = models.Metrics{
			ID:    "counter_" + strconv.Itoa(i),
			MType: models.Counter,
			Delta: &delta,
		}
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rec.Code)
		}
	}
}

// BenchmarkUpdatesHandler_MixedMetrics измеряет производительность со смешанными типами
func BenchmarkUpdatesHandler_MixedMetrics(b *testing.B) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// 50% counters, 50% gauges
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			delta := int64(i)
			metrics[i] = models.Metrics{
				ID:    "counter_" + strconv.Itoa(i),
				MType: models.Counter,
				Delta: &delta,
			}
		} else {
			value := float64(i)
			metrics[i] = models.Metrics{
				ID:    "gauge_" + strconv.Itoa(i),
				MType: models.Gauge,
				Value: &value,
			}
		}
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rec.Code)
		}
	}
}

// BenchmarkUpdatesHandler_Parallel измеряет конкурентную обработку batch-запросов
func BenchmarkUpdatesHandler_Parallel(b *testing.B) {
	storage := repository.NewMemStorage()
	handler := UpdatesHandler(storage, nil, nil)

	// Подготавливаем payload
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		delta := int64(i)
		metrics[i] = models.Metrics{
			ID:    "counter_" + strconv.Itoa(i),
			MType: models.Counter,
			Delta: &delta,
		}
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", rec.Code)
			}
		}
	})
}

// BenchmarkUpdatesHandler_MemoryUsage измеряет использование памяти
func BenchmarkUpdatesHandler_MemoryUsage(b *testing.B) {
	benchmarks := []struct {
		name       string
		numMetrics int
	}{
		{"10_metrics", 10},
		{"100_metrics", 100},
		{"1000_metrics", 1000},
		{"10000_metrics", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Подготавливаем payload
			metrics := make([]models.Metrics, bm.numMetrics)
			for i := 0; i < bm.numMetrics; i++ {
				delta := int64(i)
				metrics[i] = models.Metrics{
					ID:    "counter_" + strconv.Itoa(i),
					MType: models.Counter,
					Delta: &delta,
				}
			}

			body, err := json.Marshal(metrics)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				storage := repository.NewMemStorage()
				handler := UpdatesHandler(storage, nil, nil)

				req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)
			}
		})
	}
}

// BenchmarkUpdateMetricInStorage измеряет производительность функции updateMetricInStorage
func BenchmarkUpdateMetricInStorage(b *testing.B) {
	storage := repository.NewMemStorage()

	delta := int64(100)
	metric := models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: &delta,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := updateMetricInStorage(storage, metric)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpdateMetricInStorage_Counter vs Gauge сравнивает производительность
func BenchmarkUpdateMetricInStorage_CounterVsGauge(b *testing.B) {
	b.Run("Counter", func(b *testing.B) {
		storage := repository.NewMemStorage()
		delta := int64(1)
		metric := models.Metrics{
			ID:    "counter",
			MType: models.Counter,
			Delta: &delta,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateMetricInStorage(storage, metric)
		}
	})

	b.Run("Gauge", func(b *testing.B) {
		storage := repository.NewMemStorage()
		value := 3.14
		metric := models.Metrics{
			ID:    "gauge",
			MType: models.Gauge,
			Value: &value,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = updateMetricInStorage(storage, metric)
		}
	})
}
