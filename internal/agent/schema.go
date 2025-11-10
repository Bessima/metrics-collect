package agent

import (
	"bytes"
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"strconv"
)

type MetricRequest struct {
	metric *models.Metrics
}

func NewMetricRequest(typeMetric repository.TypeMetric, name string, value string) (*MetricRequest, error) {
	var metricRequest MetricRequest

	switch typeMetric {
	case repository.TypeCounter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return &metricRequest, err
		}
		metricRequest.metric = &models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Delta: &delta,
		}
	case repository.TypeGauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return &metricRequest, err
		}
		metricRequest.metric = &models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Value: &val,
		}
	}
	return &metricRequest, nil
}

func (metricRequest *MetricRequest) CompressJSONMetric() (*bytes.Buffer, error) {
	resp, err := json.Marshal(&metricRequest.metric)
	if err != nil {
		log.Printf("Failed to marshal metric: %v\n", err)
		return nil, err
	}

	compressData, err := Compress(resp)
	if err != nil {
		log.Printf("Failed to compress data: %v\n", err)
		return nil, err
	}

	return &compressData, nil
}
