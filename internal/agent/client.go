package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	models "github.com/Bessima/metrics-collect/internal/model"
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

func (client *Client) SendJSONMetric(typeMetric string, name string, value string) error {

	postURL := fmt.Sprintf("%s/update", client.Domain)
	requestValue := models.ShortFieldsMetric{
		ID:    name,
		Value: value,
		MType: typeMetric,
	}
	resp, err := json.Marshal(requestValue)
	if err != nil {
		log.Printf("Failed to marshal metric: %v\n", err)
		return err
	}

	response, err := client.HTTPClient.Post(postURL, "application/json", bytes.NewBuffer(resp))
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
