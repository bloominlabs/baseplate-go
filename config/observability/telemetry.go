package observability

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pyroscope-io/client/pyroscope"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bloominlabs/baseplate-go/config/env"
	"github.com/bloominlabs/baseplate-go/config/filesystem"
)

type PyroscopeLogger struct {
	logger zerolog.Logger
}

func (l *PyroscopeLogger) Infof(a string, b ...interface{})  {}
func (l *PyroscopeLogger) Debugf(a string, b ...interface{}) {}
func (l *PyroscopeLogger) Errorf(a string, b ...interface{}) {
	l.logger.Error().Msg(fmt.Sprintf(a, b...))
}

type PyroscopeConfig struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
	User  string `toml:"user"`
}

type TelemetryConfig struct {
	OTLPAddr     string `toml:"addr"`
	OTLPCAPath   string `toml:"ca_path"`
	OTLPCertPath string `toml:"cert_path"`
	OTLPKeyPath  string `toml:"key_path"`
	Insecure     bool   `toml:"insecure"`

	MetricsCollectionInterval time.Duration `toml:"metrics_collection_interval"`

	Pyroscope PyroscopeConfig `toml:"pyroscope"`

	metricsCleanup  *func()
	tracingCleanup  *func()
	profilerCleanup *func()
	watcher         *filesystem.CertificateWatcher
}

func (t *TelemetryConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&t.OTLPAddr, "otlp.addr", env.GetEnvStrDefault("OTLP_ADDR", "localhost:4317"), "hostname:port for OTLP.grpc protocol on remote OTLP receiver")
	f.StringVar(&t.OTLPCAPath, "otlp.ca.path", env.GetEnvStrDefault("OTLP_CA_PATH", ""), "Path to certificate authority used to verify outgoing OTLP receiver connections")
	f.StringVar(&t.OTLPCertPath, "otlp.cert.path", env.GetEnvStrDefault("OTLP_CERT_PATH", ""), "Path to certificate to encrypt outgoing OTLP receiver connections")
	f.StringVar(&t.OTLPKeyPath, "otlp.key.path", env.GetEnvStrDefault("OTLP_KEY_PATH", ""), "Path to private key to encrypt outgoing OTLP receiver connections")

	f.StringVar(&t.Pyroscope.URL, "pyroscope.url", env.GetEnvStrDefault("PYROSCOPE_URL", ""), "URL for uploading pyroscope traces")
	f.StringVar(&t.Pyroscope.Token, "pyroscope.token", env.GetEnvStrDefault("PYROSCOPE_TOKEN", ""), "Token used for authenticated to pyroscope")
	f.StringVar(&t.Pyroscope.User, "pyroscope.user", env.GetEnvStrDefault("PYROSCOPE_User", ""), "User used for authenticated to pyroscope")

	f.DurationVar(&t.MetricsCollectionInterval, "otlp.metrics_collection_interval", env.GetEnvDurDefault("METRICS_COLLECTION_INTERVAL", time.Minute), "User used for authenticated to pyroscope")

	f.BoolVar(&t.Insecure, "otlp.insecure", false, "Emit OTLP without needing mTLS certificate")

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
		w, err := filesystem.NewCertificateWatcher(t.OTLPCertPath, t.OTLPKeyPath, logger, time.Second*5)
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
	} else if t.Insecure {
		conf := insecure.NewCredentials()
		creds = &conf
	}

	telemetryOptions := TelemetryOptions{}
	telemetryOptions.parseOptions(options...)

	if telemetryOptions.resource == nil {
		resource, err := WithDefaultResource(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}
		telemetryOptions.resource = resource
	}

	logger.Info().Str("OTLPAddr", t.OTLPAddr).Msg("initializing observability")

	logger.Info().Str("OTLPAddr", t.OTLPAddr).Str("type", "metrics").Msg("initializing provider")
	metricOpts := WithDefaultMetricOpts(serviceName)
	metricOpts = append(metricOpts, metric.WithResource(telemetryOptions.resource))
	if len(telemetryOptions.metricOptions) > 0 {
		metricOpts = append(metricOpts, telemetryOptions.metricOptions...)
	}

	metricsCleanup, err := InitMetricsProvider(logger, t.OTLPAddr, creds, t.MetricsCollectionInterval, metricOpts...)
	if err != nil {
		return fmt.Errorf("failed to initialize metric provider %w", err)
	}
	t.metricsCleanup = &metricsCleanup
	log.Debug().Str("OTLPAddr", t.OTLPAddr).Str("type", "metrics").Msg("initialized provider")

	traceOpts := WithDefaultTracingOpts(serviceName)
	traceOpts = append(traceOpts, sdktrace.WithResource(telemetryOptions.resource))
	if len(telemetryOptions.tracingOptions) > 0 {
		traceOpts = append(traceOpts, telemetryOptions.tracingOptions...)
	}

	log.Info().Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("initializing provider")
	tracingCleanup, err := InitTraceProvider(logger, serviceName, t.OTLPAddr, creds, t.Pyroscope, traceOpts...)
	if err != nil {
		log.Fatal().Err(err).Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("failed to intialize provider")
	}
	t.tracingCleanup = &tracingCleanup
	log.Debug().Str("OTLPAddr", t.OTLPAddr).Str("type", "tracing").Msg("initialized provider")

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	if t.Pyroscope.URL != "" {
		log.Info().Str("url", t.Pyroscope.URL).Str("type", "profiling").Msg("initializing provider")

		p, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: serviceName,

			// replace this with the address of pyroscope server
			ServerAddress:     t.Pyroscope.URL,
			BasicAuthUser:     t.Pyroscope.User,
			BasicAuthPassword: t.Pyroscope.Token,

			// you can disable logging by setting this to nil
			Logger: &PyroscopeLogger{logger: log.Logger},

			UploadRate: time.Second * 60,
		})

		if err != nil {
			log.Fatal().Str("url", t.Pyroscope.URL).Str("type", "profiling").Err(err).Msg("failed to initialize provider")
		}

		cleanup := func() {
			p.Stop()
		}

		t.profilerCleanup = &cleanup
	}

	return nil
}

func (t *TelemetryConfig) Shutdown(ctx context.Context, logger zerolog.Logger) error {
	if t.metricsCleanup != nil {
		(*t.metricsCleanup)()
	}

	if t.tracingCleanup != nil {
		(*t.tracingCleanup)()
	}

	if t.profilerCleanup != nil {
		(*t.profilerCleanup)()
	}

	if t.watcher != nil {
		return t.watcher.Stop()
	}

	return nil
}
