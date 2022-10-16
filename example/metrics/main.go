package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"github.com/bloominlabs/baseplate-go/observability"
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

func init() {
	flag.StringVar(&bindPort, "bind.port", getenv("NOMAD_PORT_http", "7777"), "port for the main http server to listen on")
	flag.StringVar(&otlpAddr, "otlp.addr", getenv("OTLP_ADDR", "localhost:4317"), "hostname:port for otlp.grpc protocol on remote otlp receiver")
	flag.StringVar(&otlpCAPath, "otlp.ca.path", getenv("OTLP_CA_PATH", ""), "Path to certificate authority used to verify outgoing OTLP receiver connections")
	flag.StringVar(&otlpCertPath, "otlp.cert.path", getenv("OTLP_CERT_PATH", ""), "Path to certificate to encrypt outgoing OTLP receiver connections")
	flag.StringVar(&otlpKeyPath, "otlp.key.path", getenv("OTLP_KEY_PATH", ""), "Path to private key to encrypt outgoing OTLP receiver connections")
	flag.BoolVar(&withObservability, "with-observability", false, "Emit OTLP without needing a mTLS certificate")
}

func main() {
	flag.Parse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	log.Info().Msg("starting loadchecker")

	var creds *credentials.TransportCredentials
	if otlpCAPath != "" || otlpCertPath != "" || otlpKeyPath != "" {
		var err error
		creds, err = observability.LoadKeyPair(otlpCAPath, otlpCertPath, otlpKeyPath)

		if err != nil {
			log.Fatal().Err(err).Msg("failed to load otlp gRPC mTLS certificate")
		}
	}

	ctx := context.Background()
	if creds != nil || withObservability {
		log.Info().Str("otlpAddr", otlpAddr).Msg("initializing observability")
		resource, err := resource.New(ctx,
			resource.WithFromEnv(),
			resource.WithProcess(),
			resource.WithTelemetrySDK(),
			resource.WithHost(),
			resource.WithAttributes(
				// the service name used to display traces in backends
				semconv.ServiceNameKey.String("loadchecker"),
				// attribute.String("environment", config.Environment),
				// attribute.Int64("ID", config.ID),
			),
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create resource")
		}

		log.Info().Str("otlpAddr", otlpAddr).Str("type", "metrics").Msg("initializing provider")
		metricOpts := observability.WithDefaultMetricOpts()
		metricOpts = append(metricOpts, []metric.Option{metric.WithResource(resource)}...)
		metricsCleanup, err := observability.InitMetricsProvider(otlpAddr, creds, metricOpts...)
		if err != nil {
			log.Fatal().Err(err).Str("otlpAddr", otlpAddr).Str("type", "metrics").Msg("failed to intialize provider")
		}
		defer metricsCleanup()
		log.Debug().Str("otlpAddr", otlpAddr).Str("type", "metrics").Msg("initialized provider")

		traceOpts := observability.WithDefaultTracingOpts()
		traceOpts = append(traceOpts, []sdktrace.TracerProviderOption{sdktrace.WithResource(resource)}...)
		log.Info().Str("otlpAddr", otlpAddr).Str("type", "tracing").Msg("initializing provider")
		tracingCleanup, err := observability.InitTraceProvider(otlpAddr, creds, traceOpts...)
		if err != nil {
			log.Fatal().Err(err).Str("otlpAddr", otlpAddr).Str("type", "metrics").Msg("failed to intialize provider")
		}
		defer tracingCleanup()
		log.Debug().Str("otlpAddr", otlpAddr).Str("type", "tracing").Msg("initialized provider")
	}

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

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello")

	addr := ":" + bindPort
	log.Info().Str("addr", addr).Msg("starting http server")
	http.Handle("/", otelHandler)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to bind http server")
	}
}
