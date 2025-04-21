package middleware

import (
	"avito-backend/src/pkg/logger"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithHTTPContext(r.Context(), r.Method, r.URL.Path)

		requestID := middleware.GetReqID(r.Context())
		if requestID != "" {
			ctx = logger.WithRequestID(ctx, requestID)
		}

		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		r = r.WithContext(ctx)
		next.ServeHTTP(ww, r)

		attrs := []any{
			"duration_ms", fmt.Sprintf("%.2f", float64(time.Since(start).Milliseconds())),
			"status", ww.Status(),
			"size", ww.BytesWritten(),
		}

		slog.InfoContext(ctx, "запрос завершен", attrs...)
	})
}