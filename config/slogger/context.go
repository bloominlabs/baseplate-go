package slogger

import (
	"context"
	"io"
	"log/slog"
)

var DisabledLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// UserIDKey is the context key used to store the user ID for log enrichment.
// Set a value in the context with context.WithValue(ctx, slogger.UserIDKey{}, "user-123")
// and it will appear as a "userID" attribute in log records.
type ctxKey struct{}

// WithContext returns a copy of ctx with the receiver attached. The Logger
// attached to the provided Context (if any) will not be effected.  If the
// receiver's log level is Disabled it will only be attached to the returned
// Context if the provided Context has a previously attached Logger. If the
// provided Context has no attached Logger, a Disabled Logger will not be
// attached.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	if _, ok := ctx.Value(ctxKey{}).(*slog.Logger); !ok && logger == DisabledLogger {
		// Do not store disabled logger.
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, logger)
}

// Ctx returns the Logger associated with the ctx. If no logger
// is associated, DefaultContextLogger is returned, unless DefaultContextLogger
// is nil, in which case a disabled logger is returned.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return DisabledLogger
}
