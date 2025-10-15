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

func MainHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.New("index").Parse(htmlTemplate))
		metrics := storage.All()

		data := MetricsData{
			Title:   "System Metrics",
			Metrics: metrics,
		}

		tmpl.Execute(w, data)
	}
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 40px;
            background-color: #f0f0f0;
        }
        .container {
            width: 800px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        h1 {
            text-align: center;
            color: #333;
        }
        .metric {
            margin: 15px 0;
            padding: 10px;
            background: #f9f9f9;
            border-left: 4px solid #007bff;
        }
        .metric-name {
            font-weight: bold;
            color: #333;
        }
        .metric-value {
            font-size: 18px;
            color: #007bff;
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        {{range .Metrics}}
        <div class="metric">
            <div class="metric-name">{{.ID}}</div>
            <div class="metric-type">{{.MType}}</div>
            <div class="metric-value">{{.Delta}}</div>
            <div class="metric-value">{{.Value}}</div>
        </div>
        {{end}}
    </div>
</body>
</html>
`
