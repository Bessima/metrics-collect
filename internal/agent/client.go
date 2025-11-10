package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
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

func (client *Client) SendJSONMetric(metricRequest MetricRequest) error {
	compressData, err := metricRequest.CompressJSONMetric()
	if err != nil {
		log.Printf("Failed to compress data: %v\n", err)
		return err
	}

	postURL := fmt.Sprintf("%s/update/", client.Domain)
	req, err := http.NewRequest(http.MethodPost, postURL, compressData)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	response, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Failed to create resource at: %s and the error is: %v\n", metricRequest.metric.ID, err)
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

	log.Printf("Successful sending metric %s", metricRequest.metric.ID)
	return nil
}
