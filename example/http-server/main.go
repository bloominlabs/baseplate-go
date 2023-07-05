package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	bConfig "github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/observability"
	"github.com/bloominlabs/baseplate-go/config/server"
	bHttp "github.com/bloominlabs/baseplate-go/http"
)

const SLUG = "loadchecker"

type Config struct {
	Telemetry observability.TelemetryConfig `toml:"telemetry"`
	Server    server.ServerConfig           `toml:"server"`
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Telemetry.RegisterFlags(f)
	c.Server.RegisterFlags(f, "server")
}

func main() {
	var (
		cfg Config
	)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	_, err := bConfig.ParseConfiguration(&cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	log.Info().Msg("starting")
	if err := cfg.Telemetry.InitializeTelemetry(context.Background(), SLUG, log.Logger); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)

	mp := otel.GetMeterProvider()
	meter := mp.Meter(SLUG)
	observerLock := new(sync.RWMutex)
	underLoad := new(int64)
	labels := new([]attribute.KeyValue)

	// TODO
	gaugeObserver, err := meter.Int64ObservableGauge(
		"under_load",
		metric.WithDescription(
			"1 if the instance is 'under load'; otherwise, 0. Used to trick the autoscaler",
		),
	)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize instrument")
	}
	_, _ = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		(*observerLock).RLock()
		value := *underLoad
		labels := *labels
		(*observerLock).RUnlock()
		o.ObserveInt64(gaugeObserver, value, metric.WithAttributes(labels...))

		return nil
	})

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		logger := hlog.FromRequest(req)
		ctx := req.Context()
		_ = trace.SpanFromContext(ctx)
		_ = baggage.FromContext(ctx)

		_, _ = io.WriteString(w, "Hello, world!\n")

		if strings.Contains(req.URL.Path, "switch") {
			val, ok := req.URL.Query()["val"]
			if !ok || len(val[0]) < 1 {
				logger.Error().Str("val", val[0]).Msg("missing parameter")
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
		bHttp.OTLPHandler(SLUG),
		bHttp.HlogHandler,
		bHttp.RatelimiterMiddleware,
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.HandlerFunc(helloHandler))
	cfg.Server.UseCommonRoutes(mux, false)
	server, err := cfg.Server.NewServer(chain.Then(mux), log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}
	log.Logger.Info().Str("Addr", server.Addr).Msg("listening")

	err = server.Listen()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		server.Shutdown(ctx)
		cancel()
	}()
	if err != nil {
		log.Fatal().Err(err).Msg("error while listening to http server")
	}
}
