package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	models "github.com/Bessima/metrics-collect/internal/model"
	"go.uber.org/zap"
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

type FileStorageRepository struct {
	FileName string
}

func NewFileStorageRepository(filename string) *FileStorageRepository {
	return &FileStorageRepository{FileName: filename}
}

func (repository *FileStorageRepository) Counter(name string, value int64) error {
	metrics, err := repository.All()
	if err != nil {
		return err
	}
	hasInFile := false
	typeCounter := string(TypeCounter)

	for _, metric := range metrics {
		if metric.MType == typeCounter && metric.ID == name {
			*metric.Delta = *metric.Delta + value
			hasInFile = true
			break
		}
	}

	if !hasInFile {
		metrics = append(metrics, models.Metrics{ID: name, MType: typeCounter, Delta: &value})
	}

	return repository.Load(metrics)
}

func (repository *FileStorageRepository) ReplaceGaugeMetric(name string, value float64) error {
	metrics, err := repository.All()
	if err != nil {
		return err
	}
	hasInFile := false
	typeGauge := string(TypeGauge)

	for i := range metrics {
		if metrics[i].MType == typeGauge && metrics[i].ID == name {
			metrics[i].Value = &value
			hasInFile = true
			break
		}
	}

	if !hasInFile {
		metrics = append(metrics, models.Metrics{ID: name, MType: typeGauge, Value: &value})
	}

	return repository.Load(metrics)
}

func (repository *FileStorageRepository) GetValue(typeMetric TypeMetric, name string) (interface{}, error) {
	metric, err := repository.GetMetric(typeMetric, name)
	if err != nil {
		return nil, err
	}
	switch typeMetric {
	case TypeCounter:
		return metric.Delta, err
	case TypeGauge:
		return metric.Value, err
	default:
		err = errors.New("unknown metric type")
	}

	return nil, err
}

func (repository *FileStorageRepository) GetMetric(typeMetric TypeMetric, name string) (models.Metrics, error) {
	metrics, err := repository.All()
	if err != nil {
		return models.Metrics{}, err
	}

	typeMetricS := string(typeMetric)

	for _, metric := range metrics {
		if metric.MType == typeMetricS && metric.ID == name {
			return metric, nil
		}
	}
	err = fmt.Errorf("metric %s with type %s not found", name, typeMetricS)

	return models.Metrics{}, err
}

func (repository *FileStorageRepository) Load(metrics []models.Metrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(repository.FileName, data, 0666)
}
func (repository *FileStorageRepository) All() ([]models.Metrics, error) {
	metrics := []models.Metrics{}

	file, err := os.ReadFile(repository.FileName)
	if err != nil {
		logger.Log.Error(
			"Unable to open file",
			zap.String("filename", repository.FileName),
			zap.String("error", err.Error()),
		)
		return metrics, err
	}
	if len(file) == 0 {
		return metrics, nil
	}

	err = json.Unmarshal(file, &metrics)
	if err != nil {
		logger.Log.Error(
			"Unable to unmarshal metrics from file",
			zap.String("filename", repository.FileName),
			zap.String("error", err.Error()),
		)
		return metrics, err
	}

	return metrics, nil
}

func (repository *FileStorageRepository) Ping(ctx context.Context) error {
	err := errors.New("current command only for DB. Server is working with file storage now")
	return err
}

func (repository *FileStorageRepository) Close() error {
	return nil
}
