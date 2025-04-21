package middleware

import (
	"avito-backend/src/internal/delivery/http/middleware"
	"avito-backend/src/pkg/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		handlerStatus  int
		expectedStatus string
	}{
		{
			name:           "Success request",
			method:         "GET",
			path:           "/test",
			handlerStatus:  200,
			expectedStatus: "200",
		},
		{
			name:           "Client error",
			method:         "POST",
			path:           "/test",
			handlerStatus:  400,
			expectedStatus: "400",
		},
		{
			name:           "Server error",
			method:         "PUT",
			path:           "/test",
			handlerStatus:  500,
			expectedStatus: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
			})

			middleware := middleware.MetricsMiddleware(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			assert.Equal(t, tt.handlerStatus, rec.Code)

			metrics := metrics.HttpRequestsTotal.WithLabelValues(tt.method, tt.path, tt.expectedStatus)
			assert.NotNil(t, metrics)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name         string
		writeStatus  int
		writeBody    string
		expectStatus string
	}{
		{
			name:         "Default status",
			writeStatus:  0,
			writeBody:    "test",
			expectStatus: "200",
		},
		{
			name:         "Setting status",
			writeStatus:  404,
			writeBody:    "not found",
			expectStatus: "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := middleware.WrapResponseWriter(rec)

			if tt.writeStatus != 0 {
				wrapped.WriteHeader(tt.writeStatus)
			}

			if tt.writeBody != "" {
				_, err := wrapped.Write([]byte(tt.writeBody))
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectStatus, wrapped.Status())
		})
	}
}
