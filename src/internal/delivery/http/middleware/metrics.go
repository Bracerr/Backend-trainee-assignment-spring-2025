package middleware

import (
	"net/http"
	"strconv"
	"time"

	"avito-backend/src/pkg/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := WrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, wrapped.status).Inc()
		metrics.HttpResponseTime.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status string
}

func WrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: "200"}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = strconv.Itoa(code)
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Status() string {
	return rw.status
}
