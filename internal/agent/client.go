package agent

import (
	"fmt"
	"log"
	"net/http"
)

const Domain = "http://localhost:8080"

func SendMetric(typeMetric string, name string, value string) {
	postURL := fmt.Sprintf("%s/update/%s/%s/%s", Domain, typeMetric, name, value)
	_, err := http.Post(postURL, "text/plain; charset=utf-8", nil)
	if err != nil {
		log.Fatalf("Failed to create resource at: %s and the error is: %v\n", postURL, err)
	}

	log.Println("Successful sending: ", postURL)
}
