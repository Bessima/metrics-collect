package agent

import (
	models "github.com/Bessima/metrics-collect/internal/model"
	"testing"
)

func BenchmarkAddBaseMetricsAddBaseMetrics(b *testing.B) {
	metricsForSend := make(chan models.Metrics, 100)

	// Запускаем горутину для чтения из канала
	done := make(chan bool)
	go func() {
		for range metricsForSend {
			// Просто читаем и выбрасываем
		}
		done <- true
	}()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		AddBaseMetrics(metricsForSend, 1)
	}
	b.StopTimer()

	close(metricsForSend)
	<-done
}
