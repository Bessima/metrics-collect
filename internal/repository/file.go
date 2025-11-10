package repository

import (
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"os"
)

type MetricsFromFile struct {
	metrics  []models.Metrics
	FileName string
}

func (metrics *MetricsFromFile) UpdateMetrics(newMetrics *[]models.Metrics) error {
	metrics.metrics = *newMetrics

	data, err := json.Marshal(metrics.metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(metrics.FileName, data, 0666)
}

func (metrics *MetricsFromFile) Load() error {
	settingsFile, err := os.ReadFile(metrics.FileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(settingsFile, &metrics.metrics)
	if err != nil {
		return err
	}
	return nil
}

func (metrics *MetricsFromFile) GetMetrics() []models.Metrics {
	return metrics.metrics
}
