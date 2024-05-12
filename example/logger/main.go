package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/logger"
)

const SLUG = "acme-example"

type Config struct {
	Logger logger.LoggerConfig
}

func main() {
	var (
		cfg Config
	)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	_, err := config.ParseConfiguration(&cfg.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	logger := cfg.Logger.GetLogger()
	logger.Trace().Msg("error")
	logger.Debug().Msg("debug")
	logger.Info().Msg("info")
	logger.Warn().Msg("warn")
	logger.Error().Msg("error")
}
