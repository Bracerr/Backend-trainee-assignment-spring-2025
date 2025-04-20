package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	appMiddleware "avito-backend/src/internal/delivery/http/middleware"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		requestID      string
		handlerStatus  int
		handlerBody    string
		expectedFields map[string]interface{}
	}{
		{
			name:          "Success request with request_id",
			method:        "GET",
			path:          "/test",
			requestID:     "test-request-id",
			handlerStatus: 200,
			handlerBody:   "success",
			expectedFields: map[string]interface{}{
				"msg":         "запрос завершен",
				"status":      float64(200),
				"size":        float64(7),
				"request_id":  "test-request-id",
			},
		},
		{
			name:          "Request with error without request_id",
			method:        "POST",
			path:          "/error",
			requestID:     "",
			handlerStatus: 500,
			handlerBody:   "internal error",
			expectedFields: map[string]interface{}{
				"msg":         "запрос завершен",
				"status":      float64(500),
				"size":        float64(14),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			handler := slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			slog.SetDefault(slog.New(handler))

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				w.Write([]byte(tt.handlerBody))
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.requestID != "" {
				ctx := context.WithValue(req.Context(), middleware.RequestIDKey, tt.requestID)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()

			appMiddleware.LoggerMiddleware(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.handlerStatus, rec.Code)
			assert.Equal(t, tt.handlerBody, rec.Body.String())

			var logEntry map[string]interface{}
			err := json.Unmarshal(logBuffer.Bytes(), &logEntry)
			assert.NoError(t, err, "Лог должен быть валидным JSON")

			assert.Contains(t, logEntry, "time")
			assert.Contains(t, logEntry, "level")
			assert.Contains(t, logEntry, "duration_ms")

			for key, value := range tt.expectedFields {
				assert.Equal(t, value, logEntry[key],
					"Поле %s должно иметь значение %v, получено %v",
					key, value, logEntry[key])
			}
		})
	}
}