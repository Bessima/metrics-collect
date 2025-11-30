package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bessima/metrics-collect/internal/common"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/Bessima/metrics-collect/internal/retry"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Client struct {
	Domain     string
	HTTPClient *http.Client
}

func (client *Client) SendMetric(typeMetric string, name string, value string) error {

	postURL := fmt.Sprintf("%s/update/%s/%s/%s", client.Domain, typeMetric, name, value)
	response, err := client.HTTPClient.Post(postURL, "text/plain; charset=utf-8", nil)
	if err != nil {
		log.Printf("Failed to create resource at: %s and the error is: %v\n", postURL, err)
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v\n", err)
		}
	}()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		log.Printf("Server returned non-OK status: %d, body: %s\n", response.StatusCode, string(body))
		return fmt.Errorf("server returned status: %d", response.StatusCode)
	}

	log.Println("Successful sending: ", postURL)
	return nil
}

func (client *Client) SendData(data *bytes.Buffer, hash string) error {
	postURL := fmt.Sprintf("%s/updates/", client.Domain)

	return retry.DoRetry(context.Background(), func() error {
		req, err := http.NewRequest(http.MethodPost, postURL, data)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Encoding", "gzip")

		if hash != "" {
			req.Header.Add(common.HashHeader, hash)
		}

		var response *http.Response

		response, err = client.HTTPClient.Do(req)
		if err != nil {
			log.Printf("Failed sending resources, error is: %v\n", err)
			return err
		}

		defer func() {
			if err := response.Body.Close(); err != nil {
				log.Printf("Error closing response body: %v\n", err)
			}
		}()
		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(response.Body)
			log.Printf("Server returned non-OK status: %d, body: %s\n", response.StatusCode, string(body))
			return fmt.Errorf("server returned status: %d", response.StatusCode)
		}

		log.Print("Successful sending data")
		return nil
	}, retry.AgentRetryConfig)

}

func GetMetric(typeMetric repository.TypeMetric, name string, value string) (models.Metrics, error) {
	var metric models.Metrics

	switch typeMetric {
	case repository.TypeCounter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return metric, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Delta: &delta,
		}
	case repository.TypeGauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return metric, err
		}
		metric = models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Value: &val,
		}
	default:
		return metric, errors.New("unknown type")
	}

	return metric, nil
}

func CompressJSONMetrics(metrics []models.Metrics) (*bytes.Buffer, error) {
	resp, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v\n", err)
		return nil, err
	}

	compressData, err := Compress(resp)
	if err != nil {
		log.Printf("Failed to compress data: %v\n", err)
		return nil, err
	}

	return &compressData, nil
}
