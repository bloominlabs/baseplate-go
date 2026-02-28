package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/bloominlabs/baseplate-go/config/slogger"
)

const SLUG = "logger-example"

type Config struct {
	Logger slogger.SlogConfig `toml:"logger"`
}

func (c *Config) Validate() error {
	return c.Logger.Validate()
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Logger.RegisterFlags(f)
}

func main() {
	var cfg Config
	ctx := context.Background()

	err := config.ParseConfiguration(ctx, &cfg, nil)
	if err != nil {
		slog.Error("failed to parse configuration", "error", err)
		os.Exit(1)
	}

	logger := cfg.Logger.GetLogger(slogger.WithJSONHandler())

	logger.Log(ctx, slogger.LevelTrace, "trace message")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}
