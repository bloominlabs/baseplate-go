package slogger

import (
	"context"
	"log/slog"
)

// fanoutHandler dispatches each log record to multiple underlying handlers.
// It implements slog.Handler and is used internally by GetLogger when
// WithExtraHandler options are provided, allowing log records to be sent
// to both the primary handler (e.g., stderr) and additional handlers
// (e.g., an OTLP log exporter) simultaneously.
type fanoutHandler struct {
	handlers []slog.Handler
}

// newFanoutHandler creates a handler that fans out to all provided handlers.
// At least one handler must be provided.
func newFanoutHandler(handlers ...slog.Handler) *fanoutHandler {
	return &fanoutHandler{handlers: handlers}
}

// Enabled returns true if any of the underlying handlers are enabled for the
// given level. This ensures that records are not dropped prematurely when
// handlers have different level thresholds.
func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle dispatches the record to all underlying handlers. If any handler
// returns an error, the first error encountered is returned, but all handlers
// are still called.
func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// WithAttrs returns a new fanoutHandler where each underlying handler has
// the given attributes applied.
func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: handlers}
}

// WithGroup returns a new fanoutHandler where each underlying handler has
// the given group applied.
func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: handlers}
}
