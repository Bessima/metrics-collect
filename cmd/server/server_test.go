package main

import (
	"fmt"
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSetMetricHandler_RealRouter(t *testing.T) {
	storage := repository.NewMemStorage()

	nameCounterMetric := "testCounter"
	valueCounterMetric := int64(1)
	nameGaugeMetric := "testGauge"
	valueGaugeMetric := float64(1.1)
	storage.Counter(nameCounterMetric, valueCounterMetric)
	storage.ReplaceGaugeMetric(nameGaugeMetric, valueGaugeMetric)

	type want struct {
		code        int
		response    string
		contentType string
	}

	testServer := httptest.NewServer(getMetricRouter(&storage, &template.Template{}))
	defer testServer.Close()

	newCounterMetric := int64(3)
	newGaugeMetric := float64(3.14)

	type metric struct {
		name          string
		typeMetric    string
		value         string
		expectedValue interface{}
	}

	tests := []struct {
		name   string
		metric metric
		method string
		want   want
	}{
		{
			name: "counter metric",
			metric: metric{
				name:          nameCounterMetric,
				typeMetric:    "counter",
				value:         strconv.FormatInt(newCounterMetric, 10),
				expectedValue: valueCounterMetric + newCounterMetric,
			},
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "gauge metric",
			metric: metric{
				name:          nameCounterMetric,
				typeMetric:    "gauge",
				value:         strconv.FormatFloat(newGaugeMetric, 'f', 6, 64),
				expectedValue: newGaugeMetric,
			},
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "invalid type",
			metric: metric{
				name:          "test",
				typeMetric:    "othertype",
				value:         "1",
				expectedValue: nil,
			},
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "not digest value",
			metric: metric{
				name:          "test",
				typeMetric:    "counter",
				value:         "test",
				expectedValue: nil,
			},
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "invalid get method",
			metric: metric{
				name:          "test",
				typeMetric:    "counter",
				value:         "1",
				expectedValue: nil,
			},
			method: http.MethodGet,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "",
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/update/%s/%s/%s", tt.metric.typeMetric, tt.metric.name, tt.metric.value)
			req, err := http.NewRequest(tt.method, testServer.URL+path, nil)
			require.NoError(t, err)

			response, err := testServer.Client().Do(req)
			require.NoError(t, err)
			defer response.Body.Close()

			assert.Equal(t, tt.want.code, response.StatusCode)

			resBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))

			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))

			if tt.metric.expectedValue != nil {
				typeMetric := repository.TypeMetric(tt.metric.typeMetric)
				newValue, _ := storage.View(typeMetric, tt.metric.name)
				assert.Equal(t, tt.metric.expectedValue, newValue)
			}
		})
	}
}
