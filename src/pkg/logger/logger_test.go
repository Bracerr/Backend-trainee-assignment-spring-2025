package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHandler struct {
	enabled bool
	attrs   []slog.Attr
	record  *slog.Record
}

func (h *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabled
}

func (h *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	h.record = &r
	return nil
}

func (h *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = attrs
	return h
}

func (h *MockHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestLogContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(context.Context) context.Context
		expected LogContext
	}{
		{
			name: "With email",
			setup: func(ctx context.Context) context.Context {
				return WithEmail(ctx, "test@example.com")
			},
			expected: LogContext{Email: "test@example.com"},
		},
		{
			name: "With request_id",
			setup: func(ctx context.Context) context.Context {
				return WithRequestID(ctx, "123")
			},
			expected: LogContext{RequestID: "123"},
		},
		{
			name: "With HTTP context",
			setup: func(ctx context.Context) context.Context {
				return WithHTTPContext(ctx, "GET", "/test")
			},
			expected: LogContext{Method: "GET", Path: "/test"},
		},
		{
			name: "With PVZ ID",
			setup: func(ctx context.Context) context.Context {
				return WithPVZID(ctx, "pvz123")
			},
			expected: LogContext{PVZID: "pvz123"},
		},
		{
			name: "Combined context",
			setup: func(ctx context.Context) context.Context {
				ctx = WithEmail(ctx, "test@example.com")
				ctx = WithRequestID(ctx, "123")
				return ctx
			},
			expected: LogContext{
				Email:     "test@example.com",
				RequestID: "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup(context.Background())
			logCtx, ok := ctx.Value(logCtxKey).(LogContext)

			assert.True(t, ok)
			assert.Equal(t, tt.expected, logCtx)
		})
	}
}

func TestHandlerMiddleware(t *testing.T) {
	mock := &MockHandler{enabled: true}
	middleware := NewHandlerMiddleware(mock)

	ctx := context.Background()
	ctx = WithEmail(ctx, "test@example.com")
	ctx = WithRequestID(ctx, "123")
	ctx = WithHTTPContext(ctx, "GET", "/test")
	ctx = WithPVZID(ctx, "pvz123")

	err := middleware.Handle(ctx, slog.Record{})
	assert.NoError(t, err)

	attrMap := make(map[string]string)
	mock.record.Attrs(func(attr slog.Attr) bool {
		attrMap[attr.Key] = attr.Value.String()
		return true
	})

	assert.Equal(t, "test@example.com", attrMap["email"])
	assert.Equal(t, "123", attrMap["request_id"])
	assert.Equal(t, "GET", attrMap["method"])
	assert.Equal(t, "/test", attrMap["path"])
	assert.Equal(t, "pvz123", attrMap["pvz_id"])
}

func TestMultiHandler(t *testing.T) {
	mock1 := &MockHandler{enabled: true}
	mock2 := &MockHandler{enabled: true}

	multi := NewMultiHandler(mock1, mock2)

	err := multi.Handle(context.Background(), slog.Record{})
	assert.NoError(t, err)

	assert.NotNil(t, mock1.record)
	assert.NotNil(t, mock2.record)
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	mock1 := &MockHandler{enabled: true}
	mock2 := &MockHandler{enabled: true}
	multi := NewMultiHandler(mock1, mock2)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := multi.WithAttrs(attrs)
	newMulti, ok := newHandler.(*MultiHandler)

	assert.True(t, ok)
	assert.Len(t, newMulti.handlers, 2)

	for _, h := range []*MockHandler{mock1, mock2} {
		assert.Equal(t, attrs, h.attrs)
	}
}

func TestMultiHandler_WithGroup(t *testing.T) {
	mock1 := &MockHandler{enabled: true}
	mock2 := &MockHandler{enabled: true}
	multi := NewMultiHandler(mock1, mock2)

	groupName := "test_group"
	newHandler := multi.WithGroup(groupName)
	newMulti, ok := newHandler.(*MultiHandler)

	assert.True(t, ok)
	assert.Len(t, newMulti.handlers, 2)
}

func TestMultiHandler_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		handlers []*MockHandler
		want     bool
	}{
		{
			name: "All handlers are enabled",
			handlers: []*MockHandler{
				{enabled: true},
				{enabled: true},
			},
			want: true,
		},
		{
			name: "One handler is enabled",
			handlers: []*MockHandler{
				{enabled: false},
				{enabled: true},
			},
			want: true,
		},
		{
			name: "All handlers are disabled",
			handlers: []*MockHandler{
				{enabled: false},
				{enabled: false},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := make([]slog.Handler, len(tt.handlers))
			for i, h := range tt.handlers {
				handlers[i] = h
			}

			multi := NewMultiHandler(handlers...)
			got := multi.Enabled(context.Background(), slog.LevelInfo)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnabled(t *testing.T) {
    h := NewHandlerMiddleware(slog.NewJSONHandler(io.Discard, nil))
    assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
}

func TestWithAttrs(t *testing.T) {
    h := NewHandlerMiddleware(slog.NewJSONHandler(io.Discard, nil))
    newH := h.WithAttrs([]slog.Attr{slog.String("key", "value")})
    assert.NotNil(t, newH)
}

func TestWithGroup(t *testing.T) {
    h := NewHandlerMiddleware(slog.NewJSONHandler(io.Discard, nil))
    newH := h.WithGroup("test")
    assert.NotNil(t, newH)
}
