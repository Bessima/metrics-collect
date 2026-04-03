package main

import (
	"crypto/rsa"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/crypto_message"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
	"log"
	"net/http"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
	"github.com/Bessima/metrics-collect/internal/common"
	models "github.com/Bessima/metrics-collect/internal/model"
)

type Agent struct {
	config    *Config
	client    agent.Client
	publicKey *rsa.PublicKey
}

func NewAgent() *Agent {
	config := InitConfig()

	client := agent.Client{
		Domain:     config.getServerAddressWithProtocol(),
		HTTPClient: &http.Client{},
	}

	var pubKey *rsa.PublicKey
	if config.CryptoKey != "" {
		var err error
		pubKey, err = crypto_message.GetPublicKey(config.CryptoKey)
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}

	return &Agent{
		config:    config,
		client:    client,
		publicKey: pubKey,
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
	metricsForSend := make(chan models.Metrics, a.config.RateLimit)
	resultSending := make(chan string, a.config.RateLimit)
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
					go agent.AdditionalMemMetrics(metricsForSend)
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
	bufferData, err := agent.CompressJSONMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to compress bufferData: %v", err)
	}

	isCompressed := true
	if a.publicKey != nil {
		dataBytes := bufferData.Bytes()
		dataEncrypt, err := crypto_message.EncryptMessage(dataBytes, a.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt bufferData: %v", err)
		}
		bufferData.Reset()
		bufferData.Write(dataEncrypt)
		isCompressed = false
	}

	hash := ""
	if a.config.Key != "" {
		hash = common.GetHashData(bufferData.Bytes(), a.config.Key)
	}

	err = a.client.SendData(bufferData, hash, isCompressed)
	if err != nil {
		return fmt.Errorf("error sending metrics: %s", err)
	}

	log.Printf("metrics in count (%d) sent successfully", len(metrics))
	return nil
}
