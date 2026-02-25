package slogger

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/bloominlabs/baseplate-go/config/env"
)

const (
	// LevelTrace is a custom log level below slog.LevelDebug.
	// slog levels: DEBUG=-4, INFO=0, WARN=4, ERROR=8.
	LevelTrace = slog.Level(-8)
)

// LevelValue wraps a slog.LevelVar and implements the flag.Value interface
// so it can be used directly with flag.FlagSet for CLI parsing.
type LevelValue struct {
	*slog.LevelVar
}

// LogLevelToString converts a slog.Level to its string representation,
// handling the custom TRACE level and sub-TRACE levels.
func LogLevelToString(level slog.Level) string {
	if level < LevelTrace {
		return fmt.Sprintf("TRACE-%d", -int(level-LevelTrace))
	} else if level == LevelTrace {
		return "TRACE"
	}

	return level.String()
}

// String returns the string representation of the current level.
// If the LevelVar is nil, it defaults to INFO.
func (lv LevelValue) String() string {
	if lv.LevelVar == nil {
		return slog.LevelInfo.String()
	}

	return LogLevelToString(lv.Level())
}

// Set parses a level string (e.g., "debug", "info", "warn", "error") and sets
// the level. It implements the flag.Value interface.
func (lv LevelValue) Set(s string) error {
	return lv.UnmarshalText([]byte(s))
}

// HandlerFactory is a function that constructs a slog.Handler from an
// io.Writer and handler options. This allows users to provide custom handler
// implementations while still receiving the fully-configured HandlerOptions
// (level, ReplaceAttr for TRACE support, AddSource, etc.).
type HandlerFactory func(w io.Writer, opts *slog.HandlerOptions) slog.Handler

// Option configures the behavior of GetLogger.
type Option func(*SlogConfig)

// WithHandlerFactory sets a custom handler factory function. The factory
// receives the configured io.Writer and *slog.HandlerOptions (which includes
// the parsed log level, TRACE-aware ReplaceAttr, and AddSource setting).
//
// Use this when you need a handler other than slog.TextHandler or
// slog.JSONHandler (e.g., a third-party colored console handler).
func WithHandlerFactory(fn HandlerFactory) Option {
	return func(o *SlogConfig) {
		o.handlerFactory = fn
	}
}

// WithTextHandler configures the logger to use slog.NewTextHandler.
// This is the default handler.
func WithTextHandler() Option {
	return func(o *SlogConfig) {
		o.handlerFactory = func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewTextHandler(w, opts)
		}
	}
}

// WithJSONHandler configures the logger to use slog.NewJSONHandler.
func WithJSONHandler() Option {
	return func(o *SlogConfig) {
		o.handlerFactory = func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}
	}
}

// WithOutput sets the io.Writer for log output. Defaults to os.Stderr.
func WithOutput(w io.Writer) Option {
	return func(o *SlogConfig) {
		o.output = w
	}
}

// WithAddSource enables or disables source code location (file:line) in log
// output. Defaults to true.
func WithAddSource(b bool) Option {
	return func(o *SlogConfig) {
		o.addSource = &b
	}
}

// WithReplaceAttr sets an additional attribute replacement function. This is
// composed with the built-in TRACE level replacement — the built-in replacement
// runs first, then the user-provided function.
func WithReplaceAttr(fn func([]string, slog.Attr) slog.Attr) Option {
	return func(o *SlogConfig) {
		o.replaceAttr = fn
	}
}

// WithExtraHandler adds an additional slog.Handler that will receive all log
// records alongside the primary handler. Multiple calls append additional
// handlers. This is used to add an OTLP log exporter handler after telemetry
// initialization, so that log records are sent to both stderr and the
// collector.
//
// Internally, when extra handlers are present, GetLogger wraps the primary
// handler and all extra handlers in a fanoutHandler.
func WithExtraHandler(h slog.Handler) Option {
	return func(o *SlogConfig) {
		o.extraHandlers = append(o.extraHandlers, h)
	}
}

// SlogConfig holds configuration for creating a structured logger.
// It follows the same patterns as other config types in the baseplate-go
// project: RegisterFlags for CLI parsing, Validate, and Merge.
//
// The LogLevel field is used for TOML/flag deserialization. Internally, a
// *slog.LevelVar is maintained so that level changes via Merge or SetLevel
// take effect immediately on any logger previously returned by GetLogger,
// without needing to recreate it.
type SlogConfig struct {
	sync.RWMutex

	LogLevel string `toml:"log_level"`

	// level is the shared LevelVar passed to the handler. It is initialized
	// on the first call to GetLogger and updated by Merge/SetLevel.
	level *slog.LevelVar

	handlerFactory HandlerFactory
	output         io.Writer
	addSource      *bool // nil means use default (true)
	replaceAttr    func([]string, slog.Attr) slog.Attr
	extraHandlers  []slog.Handler
}

// LevelVar returns the underlying *slog.LevelVar used by loggers created from
// this config. It returns nil if GetLogger has not been called yet.
//
// This can be used to dynamically change the log level at runtime:
//
//	cfg.LevelVar().Set(slog.LevelWarn)
func (c *SlogConfig) LevelVar() *slog.LevelVar {
	c.RLock()
	defer c.RUnlock()
	return c.level
}

// SetLevel parses the level string and updates both the config's LogLevel
// field and the underlying LevelVar. If a logger has already been created via
// GetLogger, it will immediately respect the new level.
func (c *SlogConfig) SetLevel(s string) error {
	var l slog.Level
	if err := l.UnmarshalText([]byte(s)); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.LogLevel = s
	if c.level != nil {
		c.level.Set(l)
	}

	return nil
}

// RegisterFlags registers CLI flags for the slog configuration.
//   - -slogger.log-level: the log level (default from LOG_LEVEL env var, or "debug")
func (c *SlogConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(
		&c.LogLevel,
		"slogger.log-level",
		env.GetEnvStrDefault("LOG_LEVEL", "debug"),
		"the log level to use for the logger (trace, debug, info, warn, error)",
	)
}

// Validate checks that the configured log level is parseable.
func (c *SlogConfig) Validate() error {
	var l slog.Level
	return l.UnmarshalText([]byte(c.LogLevel))
}

// Merge applies non-zero values from o into c. If a logger has already been
// created via GetLogger, the level change takes effect immediately.
func (c *SlogConfig) Merge(o *SlogConfig) error {
	if o.LogLevel != "" {
		return c.SetLevel(o.LogLevel)
	}

	if o.addSource != nil {
		if c.addSource != nil {
			*c.addSource = *o.addSource
		} else {
			c.addSource = o.addSource
		}
	}

	if len(o.extraHandlers) > 0 {
		c.extraHandlers = o.extraHandlers
	}

	if o.handlerFactory != nil {
		c.handlerFactory = o.handlerFactory
	}

	if o.output != nil {
		c.output = o.output
	}

	if o.replaceAttr != nil {
		c.replaceAttr = o.replaceAttr
	}

	return nil
}

// initLevel ensures the shared LevelVar is initialized and synced with
// the LogLevel string. Must be called under write lock.
func (c *SlogConfig) initLevel() {
	if c.level == nil {
		c.level = &slog.LevelVar{}
	}

	if c.LogLevel != "" {
		var l slog.Level
		if err := l.UnmarshalText([]byte(c.LogLevel)); err != nil {
			// Invalid level string — default to info.
			c.level.Set(slog.LevelInfo)
		} else {
			c.level.Set(l)
		}
	}
}

// WithOptions configures the current SlogConfig to affect how future
// *slog.Logger's are returned from GetLogger.
func (c *SlogConfig) WithOptions(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

// GetLogger creates a new *slog.Logger based on the configuration and the
// provided options. The returned logger includes middleware handlers for
// OpenTelemetry trace correlation and user/server ID enrichment from context.
//
// The logger's level is backed by a shared *slog.LevelVar. Subsequent calls
// to Merge or SetLevel will update the level of all loggers previously
// returned by GetLogger without needing to recreate them.
//
// The default handler is slog.TextHandler writing to os.Stderr with source
// locations enabled. Override with WithJSONHandler(), WithTextHandler(),
// WithHandlerFactory(), WithOutput(), WithAddSource(), or WithReplaceAttr().
//
// Static attributes from NOMAD_META_user_id and NOMAD_META_server_id
// environment variables are added to all log records when set.
func (c *SlogConfig) GetLogger() *slog.Logger {
	c.Lock()
	c.initLevel()
	level := c.level
	c.Unlock()

	// Resolve output writer.
	output := c.output
	if output == nil {
		output = os.Stderr
	}

	// Resolve AddSource (default: true).
	addSource := true
	if c.addSource != nil {
		addSource = *c.addSource
	}

	// Build ReplaceAttr: handle TRACE level display, then apply user's function.
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			if lvl, ok := a.Value.Any().(slog.Level); ok {
				a.Value = slog.StringValue(LogLevelToString(lvl))
			}
		}
		if c.replaceAttr != nil {
			a = c.replaceAttr(groups, a)
		}
		return a
	}

	handlerOpts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   addSource,
		ReplaceAttr: replaceAttr,
	}

	// Resolve handler factory.
	factory := c.handlerFactory
	if factory == nil {
		factory = func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}
	}

	baseHandler := factory(output, handlerOpts)

	// If extra handlers were provided, fan out to all of them plus the primary.
	if len(c.extraHandlers) > 0 {
		all := make([]slog.Handler, 0, 1+len(c.extraHandlers))
		all = append(all, baseHandler)
		all = append(all, c.extraHandlers...)
		baseHandler = newFanoutHandler(all...)
	}

	// Wrap with middleware handlers.
	// Chain: UserInformationHandler → OTelHandler → base handler
	handler := NewUserInformationHandler(NewOTelHandler(baseHandler))

	logger := slog.New(handler)

	// Add static attributes from environment variables.
	userID := env.GetEnvStrDefault("NOMAD_META_user_id", "")
	serverID := env.GetEnvStrDefault("NOMAD_META_server_id", "")

	if userID != "" {
		logger = logger.With("userId", userID)
	}
	if serverID != "" {
		logger = logger.With("serverId", serverID)
	}

	return logger
}
