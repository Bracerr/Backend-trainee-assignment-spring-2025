package middleware

import (
	"net/http"
	"time"

	"avito-backend/src/pkg/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter для получения статуса ответа
		wrapped := wrapResponseWriter(w)
		
		// Выполняем запрос
		next.ServeHTTP(wrapped, r)

		// Записываем метрики
		duration := time.Since(start).Seconds()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, wrapped.status).Inc()
		metrics.HttpResponseTime.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status string
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: "200"}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = string(code)
	rw.ResponseWriter.WriteHeader(code)
}