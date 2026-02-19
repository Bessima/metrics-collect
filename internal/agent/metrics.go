package agent

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"

	"github.com/Bessima/metrics-collect/internal/common"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/shirou/gopsutil/v4/mem"
)

const CounterPollCountMetric = "PollCount"
const GaugeRandomMetric = "RandomValue"

func GetAllMemStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"Alloc":           m.Alloc,
		"BuckHashSys":     m.BuckHashSys,
		"Frees":           m.Frees,
		"GCCPUFraction":   m.GCCPUFraction,
		"GCSys":           m.GCSys,
		"HeapAlloc":       m.HeapAlloc,
		"HeapIdle":        m.HeapIdle,
		"HeapInuse":       m.HeapInuse,
		"HeapObjects":     m.HeapObjects,
		"HeapReleased":    m.HeapReleased,
		"HeapSys":         m.HeapSys,
		"LastGC":          m.LastGC,
		"Lookups":         m.Lookups,
		"MCacheInuse":     m.MCacheInuse,
		"MCacheSys":       m.MCacheSys,
		"MSpanInuse":      m.MSpanInuse,
		"MSpanSys":        m.MSpanSys,
		"Mallocs":         m.Mallocs,
		"NextGC":          m.NextGC,
		"NumForcedGC":     m.NumForcedGC,
		"NumGC":           m.NumGC,
		"OtherSys":        m.OtherSys,
		"PauseTotalNs":    m.PauseTotalNs,
		"StackInuse":      m.StackInuse,
		"StackSys":        m.StackSys,
		"Sys":             m.Sys,
		"TotalAlloc":      m.TotalAlloc,
		GaugeRandomMetric: rand.Int63(),
	}
}

func AddPoolCounter(metrics chan models.Metrics, poolCount int64) error {
	value, err := common.ConvertInterfaceToStr(poolCount)
	if err != nil {
		return fmt.Errorf("error converting interface to metric %s: %v", CounterPollCountMetric, err)
	}
	metric, err := GetMetric(repository.TypeCounter, CounterPollCountMetric, value)
	if err != nil {
		return fmt.Errorf("error getting object metric %s: %v", CounterPollCountMetric, err)
	}
	metrics <- metric
	return nil
}

func AddMemStats(metrics chan models.Metrics) []error {
	errors := make([]error, 0)
	for name, anyValue := range GetAllMemStats() {
		value, err := common.ConvertInterfaceToStr(anyValue)
		if err != nil {
			errors = append(
				errors,
				fmt.Errorf("error converting interface to metric %s: %v", name, err),
			)
			continue
		}
		metric, err := GetMetric(repository.TypeGauge, name, value)
		if err != nil {
			errors = append(
				errors,
				fmt.Errorf("error getting object metric %s: %v", CounterPollCountMetric, err),
			)
			continue
		}
		metrics <- metric
	}
	return errors
}

func AddBaseMetrics(metrics chan models.Metrics, poolCount int64) {
	errors := AddMemStats(metrics)
	err := AddPoolCounter(metrics, poolCount)
	if err != nil {
		errors = append(errors, err)
	}

	for _, err = range errors {
		log.Println(err)
	}
}

func AdditionalMemMetrics(metrics chan models.Metrics) {
	v, err := mem.VirtualMemory()
	if err != nil || v == nil {
		log.Printf("Error getting additional metrics %v", err)
		return
	}

	newMetrics := map[string]any{"TotalMemory": v.Total, "FreeMemory": v.Free, "CPUutilization1": v.UsedPercent}

	for name, anyValue := range newMetrics {
		value, err := common.ConvertInterfaceToStr(anyValue)
		if err != nil {
			log.Printf("Error converting interface to metric %s: %v", name, err)
			continue
		}
		metric, err := GetMetric(repository.TypeGauge, name, value)
		if err != nil {
			log.Printf("Error getting object metric %s: %v", CounterPollCountMetric, err)
			return
		}
		metrics <- metric
	}

	log.Print("Add additional metrics")

}
