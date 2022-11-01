package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/grafana/dskit/flagext"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
)

const CONFIG_FILE_FLAG = "config.file"

type Configuration interface {
	RegisterFlags(*flag.FlagSet)
}

// Parse -config.file option via separate flag set, to avoid polluting default one and calling flag.Parse on it twice.
func ParseConfigFileParameter(args []string) (configFile string) {
	// ignore errors and any output here. Any flag errors will be reported by main flag.Parse() call.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	// usage not used in these functions.
	fs.StringVar(&configFile, CONFIG_FILE_FLAG, "", "")

	// Try to find -config.file and -config.expand-env option in the flags. As Parsing stops on the first error, eg. unknown flag, we simply
	// try remaining parameters until we find config flag, or there are no params left.
	// (ContinueOnError just means that flag.Parse doesn't call panic or os.Exit, but it returns error, which we ignore)
	for len(args) > 0 {
		_ = fs.Parse(args)
		args = args[1:]
	}

	return
}

func SetupConfiguration(cfg interface{}, logger zerolog.Logger) (Watcher, error) {
	configFile := ParseConfigFileParameter(os.Args[1:])

	var watcher Watcher

	// This sets default values from flags to the config.
	// It needs to be called before parsing the config file!
	convertedCfg, ok := cfg.(Configuration)
	if !ok {
		return watcher, fmt.Errorf("could not convert passed config to 'Configuration' interface. Does it implemeted RegisterFlags correctly?")
	}
	convertedCfg.RegisterFlags(flag.CommandLine)
	if configFile != "" {
		err := ReadConfiguration(configFile, &cfg, logger)
		if err != nil {
			return watcher, fmt.Errorf("failed to read %s: %w", configFile, err)
		}

		w, err := NewRateLimitedFileWatcher([]string{configFile}, logger, time.Second*5)
		if err != nil {
			return watcher, fmt.Errorf("failed to create file watcher for %s: %w", configFile, err)
		}
		watcher = w
	}

	flagext.IgnoredFlag(flag.CommandLine, CONFIG_FILE_FLAG, "Configuration file to load.")
	flag.Parse()

	return watcher, nil
}

func ReadConfiguration(file string, config interface{}, logger zerolog.Logger) error {
	out, err := os.ReadFile(file)
	if err != nil {
		logger.Fatal().Err(err).Str("configFile", file).Msg("failed to read configuration file")
	}

	err = toml.NewDecoder(bytes.NewReader(out)).DisallowUnknownFields().Decode(config)
	if err != nil {
		var details *toml.StrictMissingError
		if !errors.As(err, &details) {
			return fmt.Errorf("err should have been a *toml.StrictMissingError, but got %s (%T)", err, err)
		}

		return fmt.Errorf("failed to decode the configuration file: \n%s", details.String())
	}

	return nil
}
