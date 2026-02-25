package slogger

import (
	"context"
	"log/slog"
)

// UserIDKey is the context key used to store the user ID for log enrichment.
// Set a value in the context with context.WithValue(ctx, slogger.UserIDKey{}, "user-123")
// and it will appear as a "userID" attribute in log records.
type UserIDKey struct{}

// ServerIDKey is the context key used to store the server ID for log enrichment.
// Set a value in the context with context.WithValue(ctx, slogger.ServerIDKey{}, "server-456")
// and it will appear as a "serverID" attribute in log records.
type ServerIDKey struct{}

// UserInformationHandler is a slog.Handler middleware that enriches log
// records with user and server identity from the context. It looks for values
// stored under UserIDKey and ServerIDKey and adds them as "userID" and
// "serverID" attributes respectively.
//
// This mirrors the behavior of the zerolog UserInformationHook in the
// config/logger package.
type UserInformationHandler struct {
	next slog.Handler
}

// NewUserInformationHandler wraps the given handler with user/server ID
// enrichment from context values.
func NewUserInformationHandler(next slog.Handler) *UserInformationHandler {
	return &UserInformationHandler{next: next}
}

// Enabled delegates to the wrapped handler.
func (h *UserInformationHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle enriches the record with userID and serverID from the context if
// present, then delegates to the wrapped handler.
func (h *UserInformationHandler) Handle(ctx context.Context, r slog.Record) error {
	if userID, ok := ctx.Value(UserIDKey{}).(string); ok {
		r.AddAttrs(slog.String("userID", userID))
	}
	if serverID, ok := ctx.Value(ServerIDKey{}).(string); ok {
		r.AddAttrs(slog.String("serverID", serverID))
	}

	return h.next.Handle(ctx, r)
}

// WithAttrs returns a new UserInformationHandler wrapping the result of
// calling WithAttrs on the underlying handler.
func (h *UserInformationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &UserInformationHandler{next: h.next.WithAttrs(attrs)}
}

// WithGroup returns a new UserInformationHandler wrapping the result of
// calling WithGroup on the underlying handler.
func (h *UserInformationHandler) WithGroup(name string) slog.Handler {
	return &UserInformationHandler{next: h.next.WithGroup(name)}
}
