package logger

import (
	"context"
	"log"
	"log/slog"
	"os"
)

type HandlerMiddleware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) slog.Handler {
	return &HandlerMiddleware{next: next}
}

func (h *HandlerMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *HandlerMiddleware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(logCtxKey).(LogContext); ok {
		if c.Email != "" {
			rec.Add("email", c.Email)
		}
		if c.RequestID != "" {
			rec.Add("request_id", c.RequestID)
		}
		if c.Method != "" {
			rec.Add("method", c.Method)
		}
		if c.Path != "" {
			rec.Add("path", c.Path)
		}
		if c.PVZID != "" {
			rec.Add("pvz_id", c.PVZID)
		}
	}
	return h.next.Handle(ctx, rec)
}

func (h *HandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddleware{next: h.next.WithAttrs(attrs)}
}

func (h *HandlerMiddleware) WithGroup(name string) slog.Handler {
	return &HandlerMiddleware{next: h.next.WithGroup(name)}
}

type LogContext struct {
	Email     string
	RequestID string
	Method    string
	Path      string
	PVZID     string
}

type contextKey int

const logCtxKey = contextKey(0)

func WithEmail(ctx context.Context, email string) context.Context {
	if c, ok := ctx.Value(logCtxKey).(LogContext); ok {
		c.Email = email
		return context.WithValue(ctx, logCtxKey, c)
	}
	return context.WithValue(ctx, logCtxKey, LogContext{Email: email})
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if c, ok := ctx.Value(logCtxKey).(LogContext); ok {
		c.RequestID = requestID
		return context.WithValue(ctx, logCtxKey, c)
	}
	return context.WithValue(ctx, logCtxKey, LogContext{RequestID: requestID})
}

func WithHTTPContext(ctx context.Context, method, path string) context.Context {
	if c, ok := ctx.Value(logCtxKey).(LogContext); ok {
		c.Method = method
		c.Path = path
		return context.WithValue(ctx, logCtxKey, c)
	}
	return context.WithValue(ctx, logCtxKey, LogContext{Method: method, Path: path})
}

func WithPVZID(ctx context.Context, pvzID string) context.Context {
	if c, ok := ctx.Value(logCtxKey).(LogContext); ok {
		c.PVZID = pvzID
		return context.WithValue(ctx, logCtxKey, c)
	}
	return context.WithValue(ctx, logCtxKey, LogContext{PVZID: pvzID})
}

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, rec slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, rec.Level) {
			if err := handler.Handle(ctx, rec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

func InitLogger() {
	file, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug})
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})

	handler := NewHandlerMiddleware(NewMultiHandler(fileHandler, stdoutHandler))
	slog.SetDefault(slog.New(handler))
}
