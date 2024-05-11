package consul

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type OpenTelemetryHook struct{}

func (h OpenTelemetryHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() || span.SpanContext().HasTraceID() {
		e.Str("traceID", span.SpanContext().TraceID().String())
		e.Str("spanID", span.SpanContext().SpanID().String())

		switch level {
		case zerolog.ErrorLevel, zerolog.PanicLevel:
			span.RecordError(fmt.Errorf(msg))
		}
	}
}

type LoggerConfig struct {
	sync.RWMutex

	LogLevel string `toml:"log_level"`
}

func (c *LoggerConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.LogLevel, "logger.log-level", env.GetEnvStrDefault("LOG_LEVEL", "debug"), "the log level to use for the logger")
}

func (c *LoggerConfig) Merge(o *LoggerConfig) error {
	if o.LogLevel != "" {
		c.LogLevel = o.LogLevel
	}

	return nil
}

func (c *LoggerConfig) GetLogger() (*zerolog.Logger, error) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Caller().Logger().Hook(OpenTelemetryHook{})

	return &logger, nil
}
