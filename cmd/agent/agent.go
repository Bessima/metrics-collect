package main

import (
	"fmt"
	"github.com/Bessima/metrics-collect/internal/agent"
	"github.com/Bessima/metrics-collect/internal/common"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
	"time"
)

type Agent struct {
	config  *Config
	client  agent.Client
	metrics map[repository.TypeMetric]map[string]any
}

func NewAgent() *Agent {
	config := InitConfig()

	metrics := agent.InitialBaseMetrics()

	client := agent.Client{
		Domain:     config.getServerAddressWithProtocol(),
		HTTPClient: &http.Client{},
	}
	return &Agent{
		config:  config,
		metrics: metrics,
		client:  client,
	}
}

func (a *Agent) sendMetrics() {
	sizeForSending := 10
	needSendMetrics := false
	var metrics []models.Metrics

	for typeMetric, metric := range a.metrics {
		for name, anyValue := range metric {
			value, err := common.ConvertInterfaceToStr(anyValue)
			if err != nil {
				log.Printf("Error converting interface to metric %s: %v", name, err)
				continue
			}
			newMetric, err := agent.GetMetric(typeMetric, name, value)
			if err != nil {
				log.Printf("Error getting object metric %s: %v", name, err)
				continue
			}
			metrics = append(metrics, newMetric)
			needSendMetrics = true

			if len(metrics) == sizeForSending {
				err := a.sendCompressMetrics(metrics)
				if err != nil {
					log.Printf("Error sending metrics: %v", err)
				}
				metrics = metrics[:0]
				needSendMetrics = false
			}
		}
	}
	if needSendMetrics {
		err := a.sendCompressMetrics(metrics)
		if err != nil {
			log.Printf("Error sending metrics: %v", err)
		}
	}
	log.Printf("All data sent successfully")
}

func (a *Agent) sendCompressMetrics(metrics []models.Metrics) error {
	data, err := agent.CompressJSONMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to compress data: %v", err)
	}
	hash := ""
	if a.config.Key != "" {
		hash = common.GetHashData(data.Bytes(), a.config.Key)
	}

	err = a.client.SendData(data, hash)
	if err != nil {
		return fmt.Errorf("error sending metrics: %s", err)
	}

	log.Printf("metrics in count (%d) sent successfully", len(metrics))
	return nil
}

func (a *Agent) Run() {
	for {

		a.sendMetrics()

		ticker := time.NewTicker(time.Duration(a.config.PoolInterval) * time.Second)
		done := make(chan bool)

		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					a.metrics = agent.UpdateMetrics(a.metrics)
				}
			}
		}()

		time.Sleep(time.Duration(a.config.ReportInterval) * time.Second)
		ticker.Stop()
		done <- true
	}
}
