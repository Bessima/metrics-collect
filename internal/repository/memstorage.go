package repository

import (
	"errors"
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
)

type TypeMetric string

const (
	TypeCounter TypeMetric = "counter"
	TypeGauge   TypeMetric = "gauge"
)

type MemoryStorage interface {
	Counter(name string, value int64)
	Replace(name string, value float64)
	View(typeMetric TypeMetric, name string) float64
}

type MemStorage struct {
	counters map[string]models.Metrics
	gauge    map[string]models.Metrics
}

func NewMemStorage() MemStorage {
	return MemStorage{
		counters: make(map[string]models.Metrics),
		gauge:    make(map[string]models.Metrics),
	}
}

func (ms *MemStorage) Counter(name string, value int64) {

	if elem, exists := ms.counters[name]; exists {
		*elem.Delta = *elem.Delta + value
		return
	}
	ms.counters[name] = models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
		Value: nil,
		Hash:  "",
	}
}

func (ms *MemStorage) ReplaceGaugeMetric(name string, value float64) {
	if elem, exists := ms.gauge[name]; exists {
		*elem.Value = value
		return
	}

	ms.gauge[name] = models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
		Delta: nil,
	}
}

func (ms *MemStorage) View(typeMetric TypeMetric, name string) (value interface{}, err error) {
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
