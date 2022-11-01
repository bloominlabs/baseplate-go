package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/grafana/dskit/flagext"
	"github.com/justinas/alice"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/propagation"
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

	Port string `toml:"port"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&c.Port, "bind.port", "", "port for the main http server to listen on")

	c.Telemetry.RegisterFlags(f)
}

func main() {
	var (
		cfg Config
	)

	configFile := bConfig.ParseConfigFileParameter(os.Args[1:])

	// This sets default values from flags to the config.
	// It needs to be called before parsing the config file!
	cfg.RegisterFlags(flag.CommandLine)

	if configFile != "" {
		out, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatal().Err(err).Str("configFile", configFile).Msg("failed to read configuration file")
		}

		err = toml.NewDecoder(bytes.NewReader(out)).DisallowUnknownFields().Decode(&cfg)
		if err != nil {
			var details *toml.StrictMissingError
			if !errors.As(err, &details) {
				panic(fmt.Sprintf("err should have been a *toml.StrictMissingError, but got %s (%T)", err, err))
			}

			panic(fmt.Sprintf("failed to decode the configuration file: \n%s", details.String()))
			// log.Fatal().Err(err).Str("configFile", configFile).Str("details", details.String()).Msg("failed to unmarshl configuration file with toml")
		}
		// TODO: read initial configuration from toml
	}

	flagext.IgnoredFlag(flag.CommandLine, bConfig.CONFIG_FILE_FLAG, "Configuration file to load.")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	log.Info().Msg("starting loadchecker")

	cfg.Telemetry.InitializeTelemetry(context.Background(), "loadchecker", log.Logger)
	defer cfg.Telemetry.Shutdown(context.Background(), log.Logger)

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
		bHttp.HlogHandler,
		bHttp.RatelimiterMiddleware,
	)

	otelHandler := otelhttp.NewHandler(
		chain.Then(http.HandlerFunc(helloHandler)),
		"Hello",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithPropagators(propagation.TraceContext{}),
	)

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("starting http server")
	http.Handle("/", otelHandler)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to bind http server")
	}
}
