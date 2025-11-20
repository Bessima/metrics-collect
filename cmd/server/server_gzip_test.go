package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/Bessima/metrics-collect/internal/agent"
	models "github.com/Bessima/metrics-collect/internal/model"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipUpdateJsonHandler(t *testing.T) {
	storage := repository.NewMemStorage()

	nameCounterMetric := "testCounter"
	valueCounterMetric := int64(1)
	nameGaugeMetric := "testGauge"
	valueGaugeMetric := float64(1.1)
	storage.Counter(nameCounterMetric, valueCounterMetric)
	storage.ReplaceGaugeMetric(nameGaugeMetric, valueGaugeMetric)
	app := NewApp(context.Background(), &storage)

	testServer := httptest.NewServer(app.getMetricRouter(&template.Template{}, nil))
	defer testServer.Close()

	newCounterMetric := int64(3)
	newGaugeMetric := float64(3.14)

	type want struct {
		code          int
		response      string
		contentType   string
		expectedValue interface{}
	}
	testsGzipResponse := []struct {
		name   string
		metric models.Metrics
		method string
		want   want
	}{
		{
			name: "gauge metric with gzip",
			metric: models.Metrics{
				ID:    nameGaugeMetric,
				MType: models.Gauge,
				Value: &newGaugeMetric,
			},
			method: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				response:      "",
				contentType:   "application/x-gzip",
				expectedValue: newGaugeMetric,
			},
		},
		{
			name: "invalid get method with gzip response",
			metric: models.Metrics{
				ID:    "test",
				MType: models.Counter,
				Delta: &newCounterMetric,
			},
			method: http.MethodGet,
			want: want{
				code:          http.StatusMethodNotAllowed,
				response:      "",
				contentType:   "application/x-gzip",
				expectedValue: nil,
			},
		},
	}
	for _, tt := range testsGzipResponse {
		t.Run(tt.name, func(t *testing.T) {
			path := testServer.URL + "/update/"
			body, _ := json.Marshal(tt.metric)
			compressBody, _ := agent.Compress(body)

			req, err := http.NewRequest(tt.method, path, &compressBody)
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			response, err := testServer.Client().Do(req)
			require.NoError(t, err)
			defer response.Body.Close()

			assert.Equal(t, tt.want.code, response.StatusCode)

			zr, err := gzip.NewReader(response.Body)

			require.NoError(t, err)

			b, err := io.ReadAll(zr)
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(b))

			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))

			if tt.want.expectedValue != nil {
				typeMetric := repository.TypeMetric(tt.metric.MType)
				newValue, _ := storage.GetValue(typeMetric, tt.metric.ID)
				assert.Equal(t, tt.want.expectedValue, newValue)
			}
		})
	}
}
