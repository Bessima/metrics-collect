package handler

import (
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
)

func updateMetricInStorage(storage repository.StorageRepositoryI, metric models.Metrics) error {
	switch repository.TypeMetric(metric.MType) {
	case repository.TypeCounter:
		if metric.Delta == nil {
			return fmt.Errorf("delta value not found for %s", metric.ID)
		}
		delta := *metric.Delta
		if err := storage.Counter(metric.ID, delta); err != nil {
			return fmt.Errorf("failed to change delta of counter metric, error: %s", err)
		}
		newValue, err := storage.GetValue(models.Counter, metric.ID)
		if err != nil {
			return fmt.Errorf("failed to get value of counter metric, error: %s", err)
		}
		log.Println("Successful counter: ", metric.ID, newValue)
		return nil
	case repository.TypeGauge:
		if metric.Value == nil {
			return fmt.Errorf("value not found for %s", metric.ID)
		}
		value := *metric.Value
		if err := storage.ReplaceGaugeMetric(metric.ID, value); err != nil {
			return fmt.Errorf("failed to change value of gauge metric, error: %s", err)
		}
		newValue, err := storage.GetValue(models.Gauge, metric.ID)
		if err != nil {
			return fmt.Errorf("failed to get value of counter metric, error: %s", err)
		}
		log.Println("Successful replacing gauge: ", metric.ID, newValue)
		return nil
	default:
		return fmt.Errorf("type %s not supported", metric.MType)
	}

}
