package repository

import (
	"context"

	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	models "github.com/Bessima/metrics-collect/internal/model"
	"go.uber.org/zap"
)

type StorageRepositoryI interface {
	Counter(name string, value int64) error
	ReplaceGaugeMetric(name string, value float64) error
	GetValue(typeMetric TypeMetric, name string) (interface{}, error)
	GetMetric(typeMetric TypeMetric, name string) (models.Metrics, error)
	Load(metrics []models.Metrics) error
	All() ([]models.Metrics, error)
	Close() error
	Ping(ctx context.Context) error
}

func UpdateMetricInFile(storage StorageRepositoryI, metricsFromFile *MetricsFromFile) {
	if metricsFromFile == nil {
		return
	}

	newMetrics, err := storage.All()
	if err != nil {
		logger.Log.Warn(err.Error())
		return
	}
	if err = metricsFromFile.UpdateMetrics(&newMetrics); err != nil {
		logger.Log.Warn(err.Error())
	} else {
		logger.Log.Info("metrics was saved in file", zap.String("path", metricsFromFile.FileName))
	}
}
