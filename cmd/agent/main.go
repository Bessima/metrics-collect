package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
)

const pollInterval = 2
const reportInterval = 10

const Domain = "http://localhost:8080"

func main() {
	metrics := agent.InitialBaseMetrics()

	for {
		start := time.Now()

		for typeMetric, m := range metrics {
			for name, anyValue := range m {
				value, err := agent.ConvertInterfaceToStr(anyValue)
				if err != nil {
					log.Printf("Error converting interface to metric %s: %v", name, err)
					continue
				}
				client := agent.Client{
					Domain:     Domain,
					HTTPClient: &http.Client{},
				}
				err = client.SendMetric(string(typeMetric), name, value)
				if err != nil {
					log.Printf("Error sending metrics: %s", err)
					continue
				}
			}
		}

		for {
			metrics = agent.UpdateMetrics(metrics)
			if time.Since(start).Seconds() >= reportInterval {
				break
			}
			time.Sleep(pollInterval * time.Second)
		}
	}

}
