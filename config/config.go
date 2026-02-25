package config

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"

	"github.com/bloominlabs/baseplate-go/config/filesystem"
)

const ConfigFileFlag = "config.file"

// Configuration is the base interface that all config types must implement.
type Configuration interface {
	RegisterFlags(*flag.FlagSet)
	Validate() error
	GetLogger() *slog.Logger
}

// WatchableConfiguration extends Configuration with a Merge method, enabling
// automatic config file watching. When ParseConfiguration receives a config
// that implements this interface AND a config file is specified, it spawns a
// background goroutine that watches the file for changes, decodes new config,
// validates it, and calls Merge.
//
// Merge should apply non-zero fields from the decoded config onto the
// receiver. The argument is a pointer to a new zero-value of the same type,
// decoded from the updated config file.
type WatchableConfiguration interface {
	Configuration
	Merge(decoded WatchableConfiguration) error
}

// ParseConfigFileParameter parses -config.file option via separate flag set, to avoid polluting default
// one and calling flag.Parse on it twice.
func ParseConfigFileParameter(args []string) (configFile string) {
	// ignore errors and any output here. Any flag errors will be reported by main flag.Parse() call.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	// usage not used in these functions.
	fs.StringVar(
		&configFile,
		ConfigFileFlag,
		"",
		"a toml configuration file to read the default configurations from. If specified, changes to this will be watched by a filesystem watcher",
	)

	// Try to find -config.file option in the flags. As Parsing stops on the
	// first error, eg. unknown flag, we simply try remaining parameters until
	// we find config flag, or there are no params left.
	for len(args) > 0 {
		_ = fs.Parse(args)
		args = args[1:]
	}

	return
}

// ParseConfiguration parses flags and an optional config file. If cfg
// implements WatchableConfiguration and a config file is specified, a
// background goroutine watches the file for changes and calls Merge
// automatically. The goroutine is canceled when ctx is canceled.
func ParseConfiguration(ctx context.Context, cfg Configuration) error {
	configFile := ParseConfigFileParameter(os.Args[1:])

	// This sets default values from flags to the config.
	// It needs to be called before parsing the config file!
	cfg.RegisterFlags(flag.CommandLine)
	if configFile != "" {
		err := DecodeConfiguration(configFile, cfg)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", configFile, err)
		}
	}

	IgnoredFlag(flag.CommandLine, ConfigFileFlag, "Configuration file to load.")
	flag.Parse()

	// If the config supports merging AND a config file was specified, start
	// a watcher goroutine.
	if wCfg, ok := cfg.(WatchableConfiguration); ok && configFile != "" {
		// NewRateLimitedFileWatcher still requires zerolog.Logger — pass Nop
		// until the filesystem package is migrated.
		w, err := filesystem.NewRateLimitedFileWatcher([]string{configFile}, zerolog.Nop(), time.Second*5)
		if err != nil {
			return fmt.Errorf("failed to create file watcher for %s: %w", configFile, err)
		}

		w.Start(ctx)

		go func() {
			for {
				logger := cfg.GetLogger()
				select {
				case <-ctx.Done():
					if err := w.Stop(); err != nil {
						logger.Error("failed to stop file watcher",
							"file", configFile,
							"error", err,
						)
					}
					return
				case event := <-w.EventsCh():
					logger.Debug("config file changed, reloading",
						"files", event.Filenames,
					)

					if err := DecodeConfiguration(configFile, wCfg); err != nil {
						logger.Error("failed to decode updated config file",
							"file", configFile,
							"error", err,
						)
						continue
					}

					if err := wCfg.Validate(); err != nil {
						logger.Error("updated config failed validation",
							"file", configFile,
							"error", err,
						)
						continue
					}

					logger.Info("config file reloaded successfully",
						"file", configFile,
					)
				}
			}
		}()
	}

	return nil
}

func DecodeConfiguration(file string, config any) error {
	out, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %w", file, err)
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
