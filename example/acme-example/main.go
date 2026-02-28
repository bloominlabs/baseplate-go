package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/observability"
	"github.com/bloominlabs/baseplate-go/config/server"
	"github.com/bloominlabs/baseplate-go/config/slogger"
	bHttp "github.com/bloominlabs/baseplate-go/http"
)

const SLUG = "acme-example"

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

	// Phase 1: Bootstrap logger.
	err := config.ParseConfiguration(ctx, &cfg, nil)
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

	chain := alice.New(
		bHttp.OTLPHandler(SLUG),
		bHttp.SlogHandler(logger),
		bHttp.RatelimiterMiddleware(),
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqLogger := slogger.FromContext(r.Context())
		reqLogger.Info("received request!")
		w.Write([]byte("Hello world"))
	})
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
