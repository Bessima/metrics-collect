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

func MainHandler(storage *repository.MemStorage, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		metrics := storage.All()

		data := MetricsData{
			Title:   "System Metrics",
			Metrics: metrics,
		}
		err := templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
