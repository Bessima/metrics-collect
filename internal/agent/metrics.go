package agent

import (
	"runtime"
	"strconv"
)

const COUNTER_CUSTOM_METRIC = "PollCount"
const GAUGE_CUSTOM_METRIC = "RandomValue"

func GetAllMemStats() map[string]string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]string{
		"Alloc":         strconv.FormatUint(m.Alloc, 10),
		"BuckHashSys":   strconv.FormatUint(m.BuckHashSys, 10),
		"Frees":         strconv.FormatUint(m.Frees, 10),
		"GCCPUFraction": strconv.FormatFloat(m.GCCPUFraction, 'f', -1, 64),
		"GCSys":         strconv.FormatUint(m.GCSys, 10),
		"HeapAlloc":     strconv.FormatUint(m.HeapAlloc, 10),
		"HeapIdle":      strconv.FormatUint(m.HeapIdle, 10),
		"HeapInuse":     strconv.FormatUint(m.HeapInuse, 10),
		"HeapObjects":   strconv.FormatUint(m.HeapObjects, 10),
		"HeapReleased":  strconv.FormatUint(m.HeapReleased, 10),
		"HeapSys":       strconv.FormatUint(m.HeapSys, 10),
		"LastGC":        strconv.FormatUint(m.LastGC, 10),
		"Lookups":       strconv.FormatUint(m.Lookups, 10),
		"MCacheInuse":   strconv.FormatUint(m.MCacheInuse, 10),
		"MCacheSys":     strconv.FormatUint(m.MCacheSys, 10),
		"MSpanInuse":    strconv.FormatUint(m.MSpanInuse, 10),
		"MSpanSys":      strconv.FormatUint(m.MSpanSys, 10),
		"Mallocs":       strconv.FormatUint(m.Mallocs, 10),
		"NextGC":        strconv.FormatUint(m.NextGC, 10),
		"NumForcedGC":   strconv.FormatUint(uint64(m.NumForcedGC), 10),
		"NumGC":         strconv.FormatUint(uint64(m.NumGC), 10),
		"OtherSys":      strconv.FormatUint(m.OtherSys, 10),
		"PauseTotalNs":  strconv.FormatUint(m.PauseTotalNs, 10),
		"StackInuse":    strconv.FormatUint(m.StackInuse, 10),
		"StackSys":      strconv.FormatUint(m.StackSys, 10),
		"Sys":           strconv.FormatUint(m.Sys, 10),
		"TotalAlloc":    strconv.FormatUint(m.TotalAlloc, 10),
	}
}
