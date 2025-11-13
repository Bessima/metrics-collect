package repository

import (
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"go.uber.org/zap"
)

func UpdateMetricInFile(storage *MemStorage, metricsFromFile *MetricsFromFile) {
	newMetrics := storage.All()
	if err := metricsFromFile.UpdateMetrics(&newMetrics); err != nil {
		logger.Log.Warn(err.Error())
	} else {
		logger.Log.Info("metrics was saved in file", zap.String("path", metricsFromFile.FileName))
	}
}
