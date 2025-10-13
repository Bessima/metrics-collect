package agent

import (
	"errors"
	"github.com/Bessima/metrics-collect/internal/repository"
	"math/rand"
	"runtime"
	"strconv"
)

const CounterPollCountMetric = "PollCount"
const GaugeRandomMetric = "RandomValue"

func GetAllMemStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"Alloc":         m.Alloc,
		"BuckHashSys":   m.BuckHashSys,
		"Frees":         m.Frees,
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         m.GCSys,
		"HeapAlloc":     m.HeapAlloc,
		"HeapIdle":      m.HeapIdle,
		"HeapInuse":     m.HeapInuse,
		"HeapObjects":   m.HeapObjects,
		"HeapReleased":  m.HeapReleased,
		"HeapSys":       m.HeapSys,
		"LastGC":        m.LastGC,
		"Lookups":       m.Lookups,
		"MCacheInuse":   m.MCacheInuse,
		"MCacheSys":     m.MCacheSys,
		"MSpanInuse":    m.MSpanInuse,
		"MSpanSys":      m.MSpanSys,
		"Mallocs":       m.Mallocs,
		"NextGC":        m.NextGC,
		"NumForcedGC":   m.NumForcedGC,
		"NumGC":         m.NumGC,
		"OtherSys":      m.OtherSys,
		"PauseTotalNs":  m.PauseTotalNs,
		"StackInuse":    m.StackInuse,
		"StackSys":      m.StackSys,
		"Sys":           m.Sys,
		"TotalAlloc":    m.TotalAlloc,
	}
}

func InitialBaseMetrics() map[repository.TypeMetric]map[string]any {
	metrics := map[repository.TypeMetric]map[string]any{}
	metrics[repository.TypeCounter] = map[string]any{CounterPollCountMetric: int64(1)}
	metrics[repository.TypeGauge] = GetAllMemStats()
	metrics[repository.TypeGauge][GaugeRandomMetric] = rand.Int63()
	return metrics
}

func UpdateMetrics(metrics map[repository.TypeMetric]map[string]any) map[repository.TypeMetric]map[string]any {
	metrics[repository.TypeGauge] = GetAllMemStats()
	metrics[repository.TypeCounter][CounterPollCountMetric] = metrics[repository.TypeCounter][CounterPollCountMetric].(int64) + 1
	metrics[repository.TypeGauge][GaugeRandomMetric] = rand.Int63()
	return metrics
}

func ConvertInterfaceToStr(anyValue any) (value string, err error) {

	switch anyValue.(type) {
	case float64:
		value = strconv.FormatFloat(anyValue.(float64), 'f', -1, 64)
	case int64:
		value = strconv.FormatInt(anyValue.(int64), 10)
	case uint32:
		value = strconv.FormatUint(uint64(anyValue.(uint32)), 10)
	case uint64:
		value = strconv.FormatUint(anyValue.(uint64), 10)
	case string:
		value = anyValue.(string)
	default:
		err = errors.New("unsupported type")
	}
	return
}
