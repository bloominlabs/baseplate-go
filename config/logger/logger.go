package logger

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type UserIDKey struct{}
type ServerIDKey struct{}

type OpenTelemetryHook struct{}

func (h OpenTelemetryHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		e.Str("traceID", span.SpanContext().TraceID().String())

		switch level {
		case zerolog.ErrorLevel, zerolog.PanicLevel:
			span.RecordError(fmt.Errorf(msg))
		case zerolog.TraceLevel:
		default:
			span.AddEvent(fmt.Sprintf("%s: %s", level.String(), msg))
		}
	}
}

type UserInformationHook struct{}

func (h UserInformationHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if userID, ok := ctx.Value(UserIDKey{}).(string); ok {
		e.Str("userID", userID)
	}
	if serverID, ok := ctx.Value(ServerIDKey{}).(string); ok {
		e.Str("serverID", serverID)
	}
}

type LoggerConfig struct {
	sync.RWMutex

	LogLevel string `toml:"log_level"`
}

func (c *LoggerConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.LogLevel, "logger.log-level", env.GetEnvStrDefault("LOG_LEVEL", "debug"), "the log level to use for the logger")
}

func (c *LoggerConfig) Validate() error {
	_, err := zerolog.ParseLevel(c.LogLevel)

	return err
}

func (c *LoggerConfig) Merge(o *LoggerConfig) error {
	if o.LogLevel != "" {
		c.LogLevel = o.LogLevel
	}

	return nil
}

func (c *LoggerConfig) GetLogger() (*zerolog.Logger, error) {
	lvl, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		return nil, err
	}

	userID := env.GetEnvStrDefault("NOMAD_META_user_id", "")
	serverID := env.GetEnvStrDefault("NOMAD_META_server_id", "")

	loggerConstructor := zerolog.New(os.Stderr).With().Timestamp().Caller()

	if userID != "" {
		loggerConstructor = loggerConstructor.Str("userId", userID)
	}
	if serverID != "" {
		loggerConstructor = loggerConstructor.Str("serverId", serverID)
	}

	logger := loggerConstructor.Logger().Hook(UserInformationHook{}).Hook(OpenTelemetryHook{}).Level(lvl)

	return &logger, nil
}
