package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/justinas/alice"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	bConfig "github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/observability"
	"github.com/bloominlabs/baseplate-go/config/server"
	"github.com/bloominlabs/baseplate-go/config/slogger"
	bHttp "github.com/bloominlabs/baseplate-go/http"
)

const SLUG = "loadchecker"

type Config struct {
	Telemetry observability.TelemetryConfig `toml:"telemetry"`
	Server    server.ServerConfig           `toml:"server"`
	Logger    slogger.SlogConfig            `toml:"logger"`
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Telemetry.RegisterFlags(f)
	c.Server.RegisterFlags(f, "server")
	c.Logger.RegisterFlags(f)
}

func main() {
	var cfg Config

	ctx := context.Background()

	// Phase 1: Bootstrap logger (stderr only).
	err := bConfig.ParseConfiguration(ctx, &cfg, nil)
	if err != nil {
		slog.Error("failed to parse configuration", "error", err)
		os.Exit(1)
	}

	logger := cfg.Logger.GetLogger(slogger.WithJSONHandler())
	ctx = slogger.NewContext(ctx, logger)

	// Phase 2: Init telemetry.
	logger.Info("starting")
	tel, err := cfg.Telemetry.InitializeTelemetry(ctx, SLUG,
		observability.WithMetrics(),
		observability.WithTracing(),
		observability.WithLogging(),
	)
	if err != nil {
		logger.Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer tel.Shutdown(ctx)

	// Phase 3: Upgrade logger with OTLP handler.
	if h := tel.SlogHandler(SLUG); h != nil {
		logger = cfg.Logger.GetLogger(
			slogger.WithJSONHandler(),
			slogger.WithExtraHandler(h),
		)
		ctx = slogger.NewContext(ctx, logger)
	}

	mp := otel.GetMeterProvider()
	meter := mp.Meter(SLUG)
	observerLock := new(sync.RWMutex)
	underLoad := new(int64)
	labels := new([]attribute.KeyValue)

	gaugeObserver, err := meter.Int64ObservableGauge(
		"under_load",
		metric.WithDescription(
			"1 if the instance is 'under load'; otherwise, 0. Used to trick the autoscaler",
		),
	)

	if err != nil {
		logger.Error("failed to initialize instrument", "error", err)
		os.Exit(1)
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
		reqLogger := slogger.FromContext(req.Context())
		_ = trace.SpanFromContext(req.Context())
		_ = baggage.FromContext(req.Context())

		_, _ = io.WriteString(w, "Hello, world!\n")

		if strings.Contains(req.URL.Path, "switch") {
			val, ok := req.URL.Query()["val"]
			if !ok || len(val[0]) < 1 {
				reqLogger.Error("missing parameter", "val", val[0])
				return
			}

			load, err := strconv.Atoi(val[0])
			if err != nil {
				reqLogger.Error("failed to convert val to int", "error", err, "val", val[0])
				return
			}

			(*observerLock).Lock()
			*underLoad = int64(load)
			(*observerLock).Unlock()
		}
	}

	chain := alice.New(
		bHttp.OTLPHandler(SLUG),
		bHttp.SlogHandler(logger),
		bHttp.RatelimiterMiddleware(),
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.HandlerFunc(helloHandler))
	cfg.Server.UseCommonRoutes(mux, false)
	srv, err := cfg.Server.NewServer(chain.Then(mux), logger)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}
	logger.Info("listening", "Addr", srv.Addr)

	err = srv.Listen()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		srv.Shutdown(shutdownCtx)
		cancel()
	}()
	if err != nil {
		logger.Error("error while listening to http server", "error", err)
		os.Exit(1)
	}
}
