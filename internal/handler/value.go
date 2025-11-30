package handler

import (
	"bytes"
	"encoding/json"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"log"
	"net/http"
)

func ValueHandler(storage repository.StorageRepositoryI) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var requestMetric models.RequestValueMetric

		var buf bytes.Buffer
		_, err := buf.ReadFrom(request.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &requestMetric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metric, err := storage.GetMetric(repository.TypeMetric(requestMetric.MType), requestMetric.ID)
		if err != nil {
			log.Println("Failed to view requestMetric metric, error: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
