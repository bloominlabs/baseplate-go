package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc/credentials"

	"github.com/bloominlabs/baseplate-go/observability"
)

type TelemetryConfig struct {
	OTLPAddr     string `toml:"addr"`
	OTLPCAPath   string `toml:"ca_path"`
	OTLPCertPath string `toml:"cert_path"`
	OTLPKeyPath  string `toml:"key_path"`
	Insecure     bool   `toml:"insecure"`

	metricsCleanup *func()
	tracingCleanup *func()
	watcher        *CertificateWatcher
}

func (t *TelemetryConfig) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&t.OTLPAddr, "otlp.addr", GetEnvStrDefault("OTLP_ADDR", "localhost:4317"), "hostname:port for OTLP.grpc protocol on remote OTLP receiver")
	flag.StringVar(&t.OTLPCAPath, "otlp.ca.path", GetEnvStrDefault("OTLP_CA_PATH", ""), "Path to certificate authority used to verify outgoing OTLP receiver connections")
	flag.StringVar(&t.OTLPCertPath, "otlp.cert.path", GetEnvStrDefault("OTLP_CERT_PATH", ""), "Path to certificate to encrypt outgoing OTLP receiver connections")
	flag.StringVar(&t.OTLPKeyPath, "otlp.key.path", GetEnvStrDefault("OTLP_KEY_PATH", ""), "Path to private key to encrypt outgoing OTLP receiver connections")
	flag.BoolVar(&t.Insecure, "otlp.insecure", false, "Emit OTLP without needing mTLS certificate")

}

type TelemetryOptions struct {
	resource       *resource.Resource
	metricOptions  []metric.Option
	tracingOptions []sdktrace.TracerProviderOption
}

type Option func(*TelemetryOptions) error

func Resource(resource resource.Resource) Option {
	return func(o *TelemetryOptions) error {
		o.resource = &resource
		return nil
	}
}

func MetricOptions(metricOptions ...metric.Option) Option {
	return func(o *TelemetryOptions) error {
		o.metricOptions = metricOptions
		return nil
	}
}

func TracingOptions(tracingOptions ...sdktrace.TracerProviderOption) Option {
	return func(o *TelemetryOptions) error {
		o.tracingOptions = tracingOptions
		return nil
	}
}

func (t *TelemetryOptions) parseOptions(opts ...Option) error {
	// Range over each options function and apply it to our API type to
	// configure it. Options functions are applied in order, with any
	// conflicting options overriding earlier calls.
	for _, option := range opts {
		err := option(t)
		if err != nil {
			return err
		}
	}

	return nil
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (t *TelemetryConfig) InitializeTelemetry(ctx context.Context, serviceName string, logger zerolog.Logger, options ...Option) error {
	var creds *credentials.TransportCredentials
	if t.OTLPCAPath != "" || t.OTLPCertPath != "" || t.OTLPKeyPath != "" {
		logger.Debug().Str("caPath", t.OTLPCAPath).Str("certPath", t.OTLPCertPath).Str("keyPath", t.OTLPKeyPath).Msg("detected mTLS credentials")
		w, err := NewCertificateWatcher(t.OTLPCertPath, t.OTLPKeyPath, logger, time.Second*5)
		if err != nil {
			return fmt.Errorf("failed to create OTLP certificate watcher: %w", err)
		}
		t.watcher = w
		_, err = t.watcher.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start certificate watcher: %w", err)
		}
		ca, err := os.ReadFile(t.OTLPCAPath)
		if err != nil {
			return fmt.Errorf("can't read ca file from %s", t.OTLPCAPath)
		}

		capool := x509.NewCertPool()
		if !capool.AppendCertsFromPEM(ca) {
			return fmt.Errorf("can't add CA cert to pool")
		}

		tlsConfig := &tls.Config{
			RootCAs:        capool,
			GetCertificate: w.GetCertificateFunc(),
		}

		conf := credentials.NewTLS(tlsConfig)
		creds = &conf
	}
	if creds == nil && !t.Insecure {
		return fmt.Errorf("'otlp.insecure' is not specified and no certificate provided")
	}

	telemetryOptions := TelemetryOptions{}
	telemetryOptions.parseOptions(options...)

	if telemetryOptions.resource == nil {
		resource, err := resource.New(ctx,
			resource.WithFromEnv(),
			resource.WithProcess(),
			resource.WithTelemetrySDK(),
			resource.WithHost(),
			resource.WithAttributes(
				// the service name used to display traces in backends
				semconv.ServiceNameKey.String(serviceName),
				// attribute.String("environment", config.Environment),
				// attribute.Int64("ID", config.ID),
			),
		)
		if err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}
		telemetryOptions.resource = resource
	}

	logger.Info().Str("OTLPAddr", t.OTLPAddr).Msg("initializing observability")

	logger.Info().Str("OTLPAddr", t.OTLPAddr).Str("type", "metrics").Msg("initializing provider")
	metricOpts := observability.WithDefaultMetricOpts()
	metricOpts = append(metricOpts, metric.WithResource(telemetryOptions.resource))
	if len(telemetryOptions.metricOptions) > 0 {
		metricOpts = append(metricOpts, telemetryOptions.metricOptions...)
	}

	metricsCleanup, err := observability.InitMetricsProvider(t.OTLPAddr, creds, metricOpts...)
	if err != nil {
		return fmt.Errorf("failed to initialize metric provider %w", err)
	}
	t.metricsCleanup = &metricsCleanup
	log.Debug().Str("OTLPAddr", t.OTLPAddr).Str("type", "metrics").Msg("initialized provider")

	traceOpts := observability.WithDefaultTracingOpts()
	traceOpts = append(traceOpts, sdktrace.WithResource(telemetryOptions.resource))
	if len(telemetryOptions.tracingOptions) > 0 {
		traceOpts = append(traceOpts, telemetryOptions.tracingOptions...)
	}

	log.Info().Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("initializing provider")
	tracingCleanup, err := observability.InitTraceProvider(t.OTLPAddr, creds, traceOpts...)
	if err != nil {
		log.Fatal().Err(err).Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("failed to intialize provider")
	}
	t.tracingCleanup = &tracingCleanup
	log.Debug().Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("initialized provider")

	return nil
}

func (t *TelemetryConfig) Shutdown(ctx context.Context, logger zerolog.Logger) error {
	if t.metricsCleanup != nil {
		(*t.metricsCleanup)()
	}

	if t.tracingCleanup != nil {
		(*t.tracingCleanup)()
	}

	if t.watcher != nil {
		return t.watcher.Stop()
	}

	return nil
}
