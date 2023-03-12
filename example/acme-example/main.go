// TODO https://github.com/cert-manager/cert-manager/issues/2131
package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/observability"
	"github.com/bloominlabs/baseplate-go/config/server"
	bHttp "github.com/bloominlabs/baseplate-go/http"
)

const SLUG = "acme-example"

type Config struct {
	Telemetry observability.TelemetryConfig
	Server    server.ServerConfig
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
	log.Info().Msg("starting")

	_, err := config.ParseConfiguration(&cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	if err := cfg.Telemetry.InitializeTelemetry(context.Background(), SLUG, log.Logger); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize telemetry")
	}
	defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)

	chain := alice.New(
		bHttp.OTLPHandler(SLUG),
		bHttp.HlogHandler,
		bHttp.RatelimiterMiddleware,
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hlog.FromRequest(r).Info().Msg("received request!")
		w.Write([]byte("Hello world"))
	})
	server, err := cfg.Server.NewServer(chain.Then(mux), log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}
	log.Logger.Info().Str("Addr", server.Addr).Msg("listening")

	err = server.Listen()
	defer server.Cleanup()
	if err != nil {
		log.Fatal().Err(err).Msg("error while listening to http server")
	}
}
