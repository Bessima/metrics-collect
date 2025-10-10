package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"runtime/metrics"
	"strconv"
	"strings"

	"github.com/Bessima/metrics-collect/internal/repository"
)

func replaceSignsInName(name string) string {
	strWithoutSlash := strings.Replace(name[1:], "/", "-", -1)
	indexForEnd := strings.Index(strWithoutSlash, ":")
	return strWithoutSlash[:indexForEnd]
}

func getMetrics() []metrics.Sample {

	samples := make([]metrics.Sample, len(metrics.All()))
	for i, m := range metrics.All() {
		samples[i].Name = m.Name
	}
	// Считываем текущее значение всех метрик
	metrics.Read(samples)

	return samples
}

func main() {
	samples := getMetrics()
	typeMetrics := [2]repository.TypeMetric{repository.TypeCounter, repository.TypeGauge}
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred: ", err)
		}
	}()

	for _, sample := range samples {
		typeMetric := string(typeMetrics[rand.Intn(len(typeMetrics))])
		if sample.Value.Kind() != metrics.KindUint64 {
			continue
		}
		value := strconv.FormatUint(sample.Value.Uint64(), 10)
		name := replaceSignsInName(sample.Name)
		postURL := "http://localhost:8080/update/" + typeMetric + "/" + name + "/" + value

		body := bytes.NewBuffer([]byte(``))
		resp, err := http.Post(postURL, "application/json; charset=utf-8", body)
		if err != nil {
			log.Fatalf("Failed to create resource at: %s and the error is: %v\n", postURL, err)
		}
		defer resp.Body.Close()

		log.Println("Successful sending: ", postURL)
	}
}
