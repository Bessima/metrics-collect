package handler

import (
	"github.com/Bessima/metrics-collect/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetMetricHandler_RealRouter(t *testing.T) {
	storage := repository.NewMemStorage()
	type want struct {
		code        int
		response    string
		contentType string
	}

	mux := http.NewServeMux()
	mux.Handle("/update/{typeMetric}/{name}/{value}", SetMetricHandler(&storage))

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "counter metric",
			url:  "/update/counter/sched-goroutines/1",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "gauge metric",
			url:  "/update/gauge/testGauge/3.14",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.url, nil)
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, req)

			assert.Equal(t, tt.want.code, response.Code)

			resBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))

			assert.Equal(t, tt.want.contentType, response.Header().Get("Content-Type"))

		})
	}
}
