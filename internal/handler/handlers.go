package handler

import (
	"github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
	"strconv"
)

func SetMetricHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		if request.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		typeMetric := request.PathValue("typeMetric")
		metric := request.PathValue("name")

		switch repository.TypeMetric(typeMetric) {
		case repository.TypeCounter:
			value, err := strconv.ParseInt(request.PathValue("value"), 10, 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.Counter(metric, value)
			log.Println("Successful counter: ", metric, storage.View(models.Counter, metric))
		case repository.TypeGauge:
			value, err := strconv.ParseFloat(request.PathValue("value"), 64)
			if err != nil {
				log.Println("Failed to parse metric value, error: ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			storage.Replace(metric, value)
			log.Println("Successful replacing gauge: ", metric, storage.View(models.Gauge, metric))
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
