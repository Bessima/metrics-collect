package main

import (
	"fmt"
	"github.com/Bessima/metrics-collect/internal/agent"
	"github.com/Bessima/metrics-collect/internal/common"
	models "github.com/Bessima/metrics-collect/internal/model"
	"log"
	"net/http"
	"time"
)

type Agent struct {
	config *Config
	client agent.Client
}

func NewAgent() *Agent {
	config := InitConfig()

	client := agent.Client{
		Domain:     config.getServerAddressWithProtocol(),
		HTTPClient: &http.Client{},
	}
	return &Agent{
		config: config,
		client: client,
	}
}

func (a *Agent) workerSendData(metrics <-chan models.Metrics, results chan<- string) {
	sizeForSending := 10
	batch := make([]models.Metrics, 0, sizeForSending)

	for metric := range metrics {
		batch = append(batch, metric)

		if len(batch) == sizeForSending {
			err := a.sendCompressMetrics(batch)
			if err != nil {
				results <- fmt.Sprintf("Error sending batch: %v", err)
			} else {
				results <- fmt.Sprintf("Batch of %d metrics sent successfully", len(batch))
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		err := a.sendCompressMetrics(batch)
		if err != nil {
			results <- fmt.Sprintf("Error sending final batch: %v", err)
		} else {
			results <- fmt.Sprintf("Final batch of %d metrics sent successfully", len(batch))
		}
	}
}

func (a *Agent) Run() {
	metricsForSend := make(chan models.Metrics, 100)
	resultSending := make(chan string, 100)
	counter := int64(1)

	defer close(metricsForSend)
	defer close(resultSending)

	for w := 0; w < a.config.RateLimit; w++ {
		go a.workerSendData(metricsForSend, resultSending)
	}

	go func() {
		for result := range resultSending {
			log.Printf("Sending result: %s", result)
		}
	}()

	for {

		ticker := time.NewTicker(time.Duration(a.config.PoolInterval) * time.Second)
		done := make(chan bool)

		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					go agent.AddBaseMetrics(metricsForSend, counter)
					counter++
				}
			}
		}()

		time.Sleep(time.Duration(a.config.ReportInterval) * time.Second)
		ticker.Stop()
		done <- true
	}

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
