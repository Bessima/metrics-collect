package main

import (
	"github.com/Bessima/metrics-collect/internal/agent"
	"github.com/Bessima/metrics-collect/internal/common"
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

	for typeMetric, m := range a.metrics {
		for name, anyValue := range m {
			value, err := common.ConvertInterfaceToStr(anyValue)
			if err != nil {
				log.Printf("Error converting interface to metric %s: %v", name, err)
				continue
			}

			err = a.client.SendMetric(string(typeMetric), name, value)
			if err != nil {
				log.Printf("Error sending metrics: %s", err)
				continue
			}
		}
	}

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
