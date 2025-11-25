package repository

import (
	"context"
	"errors"
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
	"sync"
)

type TypeMetric string

const (
	TypeCounter TypeMetric = "counter"
	TypeGauge   TypeMetric = "gauge"
)

type MemStorage struct {
	mutex    sync.RWMutex
	counters map[string]models.Metrics
	gauge    map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]models.Metrics),
		gauge:    make(map[string]models.Metrics),
	}
}

func (ms *MemStorage) Counter(name string, value int64) error {

	if elem, exists := ms.counters[name]; exists {
		ms.mutex.Lock()
		defer ms.mutex.Unlock()
		*elem.Delta = *elem.Delta + value
		return nil
	}
	ms.counters[name] = models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
		Value: nil,
		Hash:  "",
	}
	return nil
}

func (ms *MemStorage) ReplaceGaugeMetric(name string, value float64) error {
	if elem, exists := ms.gauge[name]; exists {
		ms.mutex.Lock()
		defer ms.mutex.Unlock()
		*elem.Value = value
		return nil
	}

	ms.gauge[name] = models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
		Delta: nil,
	}
	return nil
}

func (ms *MemStorage) GetValue(typeMetric TypeMetric, name string) (value interface{}, err error) {
	switch {
	case typeMetric == TypeCounter:
		if elem, exists := ms.counters[name]; exists {
			value = *elem.Delta
			return
		}
		err = errors.New("metric not found")

	case typeMetric == TypeGauge:
		if elem, exists := ms.gauge[name]; exists {
			value = *elem.Value
			return
		}
		err = errors.New("metric not found")
	default:
		err = fmt.Errorf("unknown metric type: %s", typeMetric)
	}
	return
}

func (ms *MemStorage) GetMetric(typeMetric TypeMetric, name string) (models.Metrics, error) {
	switch {
	case typeMetric == TypeCounter:
		if elem, exists := ms.counters[name]; exists {
			return elem, nil
		}
	case typeMetric == TypeGauge:
		if elem, exists := ms.gauge[name]; exists {
			return elem, nil
		}
	default:
		err := fmt.Errorf("unknown metric type: %s", typeMetric)
		return models.Metrics{}, err
	}

	err := errors.New("metric not found")
	return models.Metrics{}, err
}

func (ms *MemStorage) All() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, len(ms.counters)+len(ms.gauge))

	for _, item := range ms.counters {
		metrics = append(metrics, item)
	}
	for _, item := range ms.gauge {
		metrics = append(metrics, item)
	}
	return metrics, nil
}

func (ms *MemStorage) Load(metrics []models.Metrics) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	for _, item := range metrics {
		if TypeMetric(item.MType) == TypeCounter {
			ms.counters[item.ID] = item
		} else if TypeMetric(item.MType) == TypeGauge {
			ms.gauge[item.ID] = item
		}
	}
	return nil
}

func (repository *MemStorage) Ping(ctx context.Context) error {
	err := errors.New("Current command only for DB. Server is working with memory storage now.")
	return err
}

func (ms *MemStorage) Close() error {
	return nil
}
