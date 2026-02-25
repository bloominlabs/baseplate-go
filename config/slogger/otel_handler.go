package slogger

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// OTelHandler is a slog.Handler middleware that enriches log records with
// OpenTelemetry trace information. When a valid span exists in the context:
//   - A "traceID" attribute is added to every log record.
//   - Error and panic-level messages are recorded as span errors via
//     span.RecordError.
//   - All other levels (except TRACE) are recorded as span events.
//
// This mirrors the behavior of the zerolog OpenTelemetryHook in the
// config/logger package.
type OTelHandler struct {
	next slog.Handler
}

// NewOTelHandler wraps the given handler with OpenTelemetry trace correlation.
func NewOTelHandler(next slog.Handler) *OTelHandler {
	return &OTelHandler{next: next}
}

// Enabled delegates to the wrapped handler.
func (h *OTelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle enriches the record with trace information if a valid span is present
// in the context, then delegates to the wrapped handler.
func (h *OTelHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(slog.String("traceID", span.SpanContext().TraceID().String()))

		switch {
		case r.Level >= slog.LevelError:
			span.RecordError(fmt.Errorf("%s", r.Message))
		case r.Level > LevelTrace:
			span.AddEvent(fmt.Sprintf("%s: %s", LogLevelToString(r.Level), r.Message))
		}
		// LevelTrace and below: no span event (matches zerolog behavior).
	}

	return h.next.Handle(ctx, r)
}

// WithAttrs returns a new OTelHandler wrapping the result of calling
// WithAttrs on the underlying handler.
func (h *OTelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OTelHandler{next: h.next.WithAttrs(attrs)}
}

// WithGroup returns a new OTelHandler wrapping the result of calling
// WithGroup on the underlying handler.
func (h *OTelHandler) WithGroup(name string) slog.Handler {
	return &OTelHandler{next: h.next.WithGroup(name)}
}
