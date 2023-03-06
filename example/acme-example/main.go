// TODO https://github.com/cert-manager/cert-manager/issues/2131
package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/observability"
	"github.com/bloominlabs/baseplate-go/config/server"
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

	if cfg.Telemetry.OTLPCAPath != "" || cfg.Telemetry.OTLPCertPath != "" || cfg.Telemetry.OTLPKeyPath != "" || cfg.Telemetry.Insecure {
		if err := cfg.Telemetry.InitializeTelemetry(context.Background(), SLUG, log.Logger); err != nil {
			log.Fatal().Err(err).Msg("failed to initialize telemetry")
		}
		defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})
	server, err := cfg.Server.NewServer(mux, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}
	log.Logger.Info().Str("Addr", server.Addr).Msg("listening")

	if server.TLSConfig != nil {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatal().Err(err).Msg("error while starting http server")
		}
	} else {
		log.Warn().Msg("running http server without https. this is not allowed in production")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("error while starting http server")
		}
	}
}
