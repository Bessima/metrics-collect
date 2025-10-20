package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMetric(t *testing.T) {
	type args struct {
		typeMetric string
		name       string
		value      string
	}
	tests := []struct {
		name          string
		args          args
		handler       http.HandlerFunc
		expectedError bool
	}{
		{
			name: "success case",
			args: args{
				typeMetric: "type",
				name:       "name",
				value:      "",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedError: false,
		},
		{
			name: "server error",
			args: args{
				typeMetric: "type",
				name:       "name",
				value:      "",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := Client{
				HTTPClient: server.Client(),
				Domain:     server.URL,
			}
			if err := client.SendMetric(tt.args.typeMetric, tt.args.name, tt.args.value); (err != nil) != tt.expectedError {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.expectedError)
			}
		})
	}
}
