package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/trace"

	bConfig "github.com/bloominlabs/baseplate-go/config"
	bHttp "github.com/bloominlabs/baseplate-go/http"
)

var otlpAddr string
var otlpCAPath string
var otlpCertPath string
var otlpKeyPath string
var bindPort string

var withObservability bool

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok == true {
		return val
	}

	return def
}

type Config struct {
	Telemetry bConfig.TelemetryConfig `toml:"telemetry"`

	Port              string `toml:"port"`
	WithObservability bool   `toml:"with_observability"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Port, "bind.port", "8080", "port for the main http server to listen on")
	f.BoolVar(&c.WithObservability, "with-observability", false, "Emit OTLP without needing a mTLS certificate")

	c.Telemetry.RegisterFlags(f)
}

func main() {
	var (
		cfg Config
	)

	_, err := bConfig.ParseConfiguration(&cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	cfg.Telemetry.InitializeTelemetry(context.Background(), "serverd", log.Logger)
	defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	log.Info().Msg("starting loadchecker")

	if cfg.Telemetry.OTLPCAPath != "" || cfg.Telemetry.OTLPCertPath != "" || cfg.Telemetry.OTLPKeyPath != "" || cfg.WithObservability {
		if err := cfg.Telemetry.InitializeTelemetry(context.Background(), "loadchecker", log.Logger); err != nil {
			log.Fatal().Err(err).Msg("failed to initialize telemetry")
		}
		defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)
	}

	mp := global.MeterProvider()
	meter := mp.Meter("loadchecker")
	observerLock := new(sync.RWMutex)
	underLoad := new(int64)
	labels := new([]attribute.KeyValue)

	gaugeObserver, err := meter.AsyncInt64().Gauge(
		"under_load",
		instrument.WithDescription(
			"1 if the instance is 'under load'; otherwise, 0. Used to trick the autoscaler",
		),
	)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize instrument")
	}
	_ = meter.RegisterCallback([]instrument.Asynchronous{gaugeObserver}, func(ctx context.Context) {
		(*observerLock).RLock()
		value := *underLoad
		labels := *labels
		(*observerLock).RUnlock()
		gaugeObserver.Observe(ctx, value, labels...)
	})

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		log := hlog.FromRequest(req)
		ctx := req.Context()
		_ = trace.SpanFromContext(ctx)
		_ = baggage.FromContext(ctx)

		_, _ = io.WriteString(w, "Hello, world!\n")

		if strings.Contains(req.URL.Path, "switch") {
			val, ok := req.URL.Query()["val"]
			if !ok || len(val[0]) < 1 {
				log.Error().Str("val", val[0]).Msg("missing parameter")
				return
			}

			load, err := strconv.Atoi(val[0])
			if err != nil {
				log.Error().Err(err).Str("val", val[0]).Msg("failed to convert val to int")
				return
			}

			(*observerLock).Lock()
			*underLoad = int64(load)
			(*observerLock).Unlock()
		}
	}

	chain := alice.New(
		bHttp.OTLPHandler("loadchecker"),
		bHttp.TraceIDHandler("traceID"),
		bHttp.HlogHandler,
		bHttp.RatelimiterMiddleware,
	)

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting http server")
	http.Handle("/", chain.Then(http.HandlerFunc(helloHandler)))
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to bind http server")
	}
}
