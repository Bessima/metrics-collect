package main

import (
	"log"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
)

const pollInterval = 2
const reportInterval = 10

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
				agent.SendMetric(string(typeMetric), name, value)
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
