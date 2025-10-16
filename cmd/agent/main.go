package main

import (
	"github.com/Bessima/metrics-collect/internal/common"
	"log"
	"net/http"
	"time"

	"github.com/Bessima/metrics-collect/internal/agent"
)

func main() {
	flags := Flags{}
	flags.Init()

	metrics := agent.InitialBaseMetrics()

	for {
		start := time.Now()

		for typeMetric, m := range metrics {
			for name, anyValue := range m {
				value, err := common.ConvertInterfaceToStr(anyValue)
				if err != nil {
					log.Printf("Error converting interface to metric %s: %v", name, err)
					continue
				}
				client := agent.Client{
					Domain:     flags.getServerAddressWithProtocol(),
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
			if time.Since(start).Seconds() >= flags.getReportInterval() {
				break
			}
			time.Sleep(flags.getPollInterval() * time.Second)
		}
	}

}
