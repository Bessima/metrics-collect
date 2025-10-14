package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const Domain = "http://localhost:8080"

func SendMetric(typeMetric string, name string, value string) error {
	client := &http.Client{}

	postURL := fmt.Sprintf("%s/update/%s/%s/%s", Domain, typeMetric, name, value)
	response, err := client.Post(postURL, "text/plain; charset=utf-8", nil)
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
