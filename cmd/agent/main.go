package main

import (
	"fmt"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
)

const pollInterval = 2
const reportInterval = 10

func main() {
	metrics := agent.GetAllMemStats()

	for {
		start := time.Now()

		for name, value := range metrics {
			typeMetric := string(repository.TypeGauge)

			postURL := fmt.Sprintf("http://localhost:8080/update/%s/%s/%s", typeMetric, name, value)
			resp, err := http.Post(postURL, "text/plain; charset=utf-8", nil)
			if err != nil {
				log.Fatalf("Failed to create resource at: %s and the error is: %v\n", postURL, err)
			}
			defer resp.Body.Close()

			log.Println("Successful sending: ", postURL)
		}

		for {
			metrics = agent.GetAllMemStats()

			if time.Now().Sub(start).Seconds() >= reportInterval {
				break
			}

			time.Sleep(pollInterval * time.Second)
		}
	}

}
