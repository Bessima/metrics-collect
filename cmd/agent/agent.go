package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
	"github.com/Bessima/metrics-collect/internal/common"
	"github.com/Bessima/metrics-collect/internal/cryptomessage"
	"github.com/Bessima/metrics-collect/internal/middlewares/logger"
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
		pubKey, err = cryptomessage.GetPublicKey(config.CryptoKey)
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
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	metricsForSend := make(chan models.Metrics, a.config.RateLimit)
	resultSending := make(chan string, a.config.RateLimit)
	counter := int64(1)

	// Track workers so we can wait for them to finish sending before exit.
	var workerWg sync.WaitGroup
	for w := 0; w < a.config.RateLimit; w++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			a.workerSendData(metricsForSend, resultSending)
		}()
	}

	// Drain result log in background; exits when resultSending is closed.
	var loggerWg sync.WaitGroup
	loggerWg.Add(1)
	go func() {
		defer loggerWg.Done()
		for result := range resultSending {
			log.Printf("Sending result: %s", result)
		}
	}()

	// Track in-flight collector goroutines so we don't close the channel
	// while they are still writing to it.
	var collectorWg sync.WaitGroup

	ticker := time.NewTicker(time.Duration(a.config.PoolInterval) * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-rootCtx.Done():
			fmt.Println("Received shutdown signal, shutting down.")
			logger.Log.Warn("Received shutdown signal, shutting down.")
			break loop
		case <-ticker.C:
			cnt := counter
			collectorWg.Add(2)
			go func() {
				defer collectorWg.Done()
				agent.AddBaseMetrics(metricsForSend, cnt)
			}()
			go func() {
				defer collectorWg.Done()
				agent.AdditionalMemMetrics(metricsForSend)
			}()
			counter++
		}
	}

	// Wait for all in-flight collectors to finish writing to the channel.
	collectorWg.Wait()

	// Signal workers that there are no more metrics.
	close(metricsForSend)

	// Wait for all workers to finish draining and sending remaining data.
	workerWg.Wait()

	// Signal logger and wait for it to flush.
	close(resultSending)
	loggerWg.Wait()
}

func (a *Agent) sendCompressMetrics(metrics []models.Metrics) error {
	bufferData, err := agent.CompressJSONMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to compress bufferData: %v", err)
	}

	isCompressed := true
	if a.publicKey != nil {
		dataBytes := bufferData.Bytes()
		dataEncrypt, err := cryptomessage.EncryptMessage(dataBytes, a.publicKey)
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
