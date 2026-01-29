package handler

import (
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"html/template"
	"net/http"
)

type MetricsData struct {
	Title   string
	Metrics []models.Metrics
}

func MainHandler(storage repository.StorageRepositoryI, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		metrics, err := storage.All()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		data := MetricsData{
			Title:   "System Metrics",
			Metrics: metrics,
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if templates != nil {
			err = templates.ExecuteTemplate(w, "index.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}

	}
}
