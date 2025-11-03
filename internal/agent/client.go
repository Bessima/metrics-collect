package agent

import (
	"encoding/json"
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
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

func (client *Client) SendJSONMetric(typeMetric repository.TypeMetric, name string, value string) error {
	postURL := fmt.Sprintf("%s/update/", client.Domain)
	var requestValue models.Metrics

	switch typeMetric {
	case repository.TypeCounter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		requestValue = models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Delta: &delta,
		}
	case repository.TypeGauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		requestValue = models.Metrics{
			ID:    name,
			MType: string(typeMetric),
			Value: &val,
		}
	}

	resp, err := json.Marshal(requestValue)
	if err != nil {
		log.Printf("Failed to marshal metric: %v\n", err)
		return err
	}

	compressData, err := Compress(resp)
	if err != nil {
		log.Printf("Failed to compress data: %v\n", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, postURL, &compressData)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	response, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Failed to create resource at: %s and the error is: %v\n", resp, err)
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

	log.Printf("Successful sending metric %s", resp)
	return nil
}
