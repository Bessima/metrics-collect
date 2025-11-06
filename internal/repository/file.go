package repository

import (
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"os"
)

type MetricsFromFile struct {
	Metrics []models.Metrics
}

func (metrics MetricsFromFile) Save(fname string) error {
	data, err := json.Marshal(metrics.Metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0666)
}

func (metrics *MetricsFromFile) Load(fname string) error {
	settingsFile, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	err = json.Unmarshal(settingsFile, &metrics.Metrics)
	if err != nil {
		return err
	}
	return nil
}
