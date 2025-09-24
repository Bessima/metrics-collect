package repository

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
	counters map[string]int64
	gauge    map[string]float64
}

func NewMemStorage() MemStorage {
	return MemStorage{
		counters: make(map[string]int64),
		gauge:    make(map[string]float64),
	}
}

func (ms *MemStorage) Counter(name string, value int64) {
	_, exists := ms.counters[name]
	if !exists {
		ms.counters[name] = value
	} else {
		ms.counters[name] = ms.counters[name] + value
	}
}

func (ms *MemStorage) Replace(name string, value float64) {
	ms.gauge[name] = value
}

func (ms *MemStorage) View(typeMetric TypeMetric, name string) interface{} {
	switch {
	case typeMetric == TypeCounter:
		return ms.counters[name]
	case typeMetric == TypeGauge:

		return ms.gauge[name]
	}
	return nil
}
