package observability

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/grafana/pyroscope-go"
	"github.com/rs/zerolog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bloominlabs/baseplate-go/config/env"
	"github.com/bloominlabs/baseplate-go/config/filesystem"
	"github.com/bloominlabs/baseplate-go/config/slogger"
)

// pyroscopeLogger adapts slog to pyroscope's logger interface.
type pyroscopeLogger struct {
	logger *slog.Logger
}

func (l *pyroscopeLogger) Infof(a string, b ...interface{})  {}
func (l *pyroscopeLogger) Debugf(a string, b ...interface{}) {}
func (l *pyroscopeLogger) Errorf(a string, b ...interface{}) {
	l.logger.Error(fmt.Sprintf(a, b...))
}

// PyroscopeConfig holds configuration for the Pyroscope continuous profiler.
type PyroscopeConfig struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
	User  string `toml:"user"`
}

// TelemetryConfig holds the configuration for OTLP telemetry. It provides
// flag registration and config file parsing. Use InitializeTelemetry to
// create a Telemetry handle from this config.
type TelemetryConfig struct {
	OTLPAddr     string `toml:"addr"`
	OTLPCAPath   string `toml:"ca_path"`
	OTLPCertPath string `toml:"cert_path"`
	OTLPKeyPath  string `toml:"key_path"`
	Insecure     bool   `toml:"insecure"`

	ServiceName string `toml:"service_name"`

	MetricsCollectionInterval time.Duration `toml:"metrics_collection_interval"`

	Pyroscope PyroscopeConfig `toml:"pyroscope"`
}

func (t *TelemetryConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&t.OTLPAddr, "otlp.addr", env.GetEnvStrDefault("OTLP_ADDR", "localhost:4317"), "hostname:port for OTLP.grpc protocol on remote OTLP receiver")
	f.StringVar(&t.OTLPCAPath, "otlp.ca.path", env.GetEnvStrDefault("OTLP_CA_PATH", ""), "Path to certificate authority used to verify outgoing OTLP receiver connections")
	f.StringVar(&t.OTLPCertPath, "otlp.cert.path", env.GetEnvStrDefault("OTLP_CERT_PATH", ""), "Path to certificate to encrypt outgoing OTLP receiver connections")
	f.StringVar(&t.OTLPKeyPath, "otlp.key.path", env.GetEnvStrDefault("OTLP_KEY_PATH", ""), "Path to private key to encrypt outgoing OTLP receiver connections")

	f.StringVar(&t.Pyroscope.URL, "pyroscope.url", env.GetEnvStrDefault("PYROSCOPE_URL", ""), "URL for uploading pyroscope traces")
	f.StringVar(&t.Pyroscope.Token, "pyroscope.token", env.GetEnvStrDefault("PYROSCOPE_TOKEN", ""), "Token used for authenticated to pyroscope")
	f.StringVar(&t.Pyroscope.User, "pyroscope.user", env.GetEnvStrDefault("PYROSCOPE_USER", ""), "User used for authenticated to pyroscope")

	f.StringVar(&t.ServiceName, "service-name", env.GetEnvStrDefault("SERVICE_NAME", ""), "Service name to use for telemetry")

	f.DurationVar(&t.MetricsCollectionInterval, "otlp.metrics_collection_interval", env.GetEnvDurDefault("METRICS_COLLECTION_INTERVAL", time.Minute), "Interval between metrics collections")

	f.BoolVar(&t.Insecure, "otlp.insecure", false, "Emit OTLP without needing mTLS certificate")
}

func (t *TelemetryConfig) Merge(o *TelemetryConfig) error {
	if o.Pyroscope.URL != "" {
		t.Pyroscope.URL = o.Pyroscope.URL
	}
	if o.Pyroscope.Token != "" {
		t.Pyroscope.Token = o.Pyroscope.Token
	}
	if o.Pyroscope.User != "" {
		t.Pyroscope.User = o.Pyroscope.User
	}
	return nil
}

func (t *TelemetryConfig) Validate() error {
	return nil
}

// ---------------------------------------------------------------------------
// Per-signal option types
// ---------------------------------------------------------------------------

type metricsConfig struct {
	extraOpts          []metric.Option
	collectionInterval time.Duration // 0 = use TelemetryConfig default
}

type tracingConfig struct {
	extraOpts []sdktrace.TracerProviderOption
}

type profilingConfig struct {
	pyroscope PyroscopeConfig
}

type loggingConfig struct {
	extraOpts []sdklog.LoggerProviderOption
}

// TelemetryOptions aggregates per-signal configs. nil fields mean the signal
// is disabled.
type TelemetryOptions struct {
	resource  *resource.Resource
	metrics   *metricsConfig
	tracing   *tracingConfig
	profiling *profilingConfig
	logging   *loggingConfig
}

// Option configures which telemetry signals to enable.
type Option func(*TelemetryOptions) error

// WithResource sets a custom OTel resource. If not provided,
// WithDefaultResource is used.
func WithResource(r *resource.Resource) Option {
	return func(o *TelemetryOptions) error {
		o.resource = r
		return nil
	}
}

// WithMetrics enables the metrics signal. Additional metric.Option values are
// appended to the default provider options.
func WithMetrics(opts ...metric.Option) Option {
	return func(o *TelemetryOptions) error {
		if o.metrics == nil {
			o.metrics = &metricsConfig{}
		}
		o.metrics.extraOpts = append(o.metrics.extraOpts, opts...)
		return nil
	}
}

// WithMetricsCollectionInterval overrides the metrics collection interval.
// Must be combined with WithMetrics.
func WithMetricsCollectionInterval(d time.Duration) Option {
	return func(o *TelemetryOptions) error {
		if o.metrics == nil {
			o.metrics = &metricsConfig{}
		}
		o.metrics.collectionInterval = d
		return nil
	}
}

// WithTracing enables the tracing signal. Additional TracerProviderOption
// values are appended to the default provider options.
func WithTracing(opts ...sdktrace.TracerProviderOption) Option {
	return func(o *TelemetryOptions) error {
		if o.tracing == nil {
			o.tracing = &tracingConfig{}
		}
		o.tracing.extraOpts = append(o.tracing.extraOpts, opts...)
		return nil
	}
}

// WithProfiling enables continuous profiling via Pyroscope. The PyroscopeConfig
// fields override those from TelemetryConfig.Pyroscope when non-empty.
func WithProfiling(cfg PyroscopeConfig) Option {
	return func(o *TelemetryOptions) error {
		o.profiling = &profilingConfig{pyroscope: cfg}
		return nil
	}
}

// WithLogging enables the OTLP log signal. Additional LoggerProviderOption
// values are appended to the defaults.
func WithLogging(opts ...sdklog.LoggerProviderOption) Option {
	return func(o *TelemetryOptions) error {
		if o.logging == nil {
			o.logging = &loggingConfig{}
		}
		o.logging.extraOpts = append(o.logging.extraOpts, opts...)
		return nil
	}
}

func (o *TelemetryOptions) parseOptions(opts ...Option) error {
	for _, option := range opts {
		if err := option(o); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Telemetry handle
// ---------------------------------------------------------------------------

// Telemetry holds the initialized providers and exposes a unified Shutdown.
type Telemetry struct {
	meterProvider  *metric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
	logProvider    *sdklog.LoggerProvider
	profiler       *pyroscope.Profiler
	watcher        *filesystem.CertificateWatcher
}

// MeterProvider returns the initialized MeterProvider, or nil if metrics were
// not enabled.
func (t *Telemetry) MeterProvider() *metric.MeterProvider {
	if t == nil {
		return nil
	}
	return t.meterProvider
}

// TracerProvider returns the initialized TracerProvider, or nil if tracing was
// not enabled.
func (t *Telemetry) TracerProvider() *sdktrace.TracerProvider {
	if t == nil {
		return nil
	}
	return t.tracerProvider
}

// LogProvider returns the initialized LoggerProvider, or nil if logging was
// not enabled.
func (t *Telemetry) LogProvider() *sdklog.LoggerProvider {
	if t == nil {
		return nil
	}
	return t.logProvider
}

// SlogHandler returns an otelslog bridge handler backed by the Telemetry's
// LoggerProvider. Returns nil if logging was not enabled.
func (t *Telemetry) SlogHandler(name string) slog.Handler {
	if t == nil || t.logProvider == nil {
		return nil
	}
	return NewOTLPSlogHandler(name, t.logProvider)
}

// Shutdown gracefully shuts down all initialized providers. Errors from
// individual shutdowns are joined via errors.Join.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil {
		return nil
	}
	var errs []error

	if t.logProvider != nil {
		if err := t.logProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("log provider shutdown: %w", err))
		}
	}
	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}
	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}
	if t.profiler != nil {
		if err := t.profiler.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("profiler shutdown: %w", err))
		}
	}
	if t.watcher != nil {
		if err := t.watcher.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("certificate watcher shutdown: %w", err))
		}
	}

	return errors.Join(errs...)
}

// ---------------------------------------------------------------------------
// InitializeTelemetry
// ---------------------------------------------------------------------------

// InitializeTelemetry creates providers for the enabled signals and returns a
// Telemetry handle. When no With*() options are provided, nothing is
// initialized and an empty (but non-nil) Telemetry is returned.
//
// The logger used for init-time messages is retrieved from ctx via
// slogger.FromContext.
func (t *TelemetryConfig) InitializeTelemetry(ctx context.Context, serviceName string, options ...Option) (*Telemetry, error) {
	logger := slogger.FromContext(ctx)

	if t.ServiceName == "" {
		t.ServiceName = serviceName
	}

	telOpts := TelemetryOptions{}
	if err := telOpts.parseOptions(options...); err != nil {
		return nil, fmt.Errorf("failed to parse telemetry options: %w", err)
	}

	// Resolve resource.
	if telOpts.resource == nil {
		r, err := WithDefaultResource(ctx, t.ServiceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}
		telOpts.resource = r
	}

	// Resolve gRPC credentials.
	var creds *credentials.TransportCredentials
	tel := &Telemetry{}

	if t.OTLPCAPath != "" || t.OTLPCertPath != "" || t.OTLPKeyPath != "" {
		logger.Debug("detected mTLS credentials",
			"caPath", t.OTLPCAPath,
			"certPath", t.OTLPCertPath,
			"keyPath", t.OTLPKeyPath,
		)
		// CertificateWatcher still takes zerolog.Logger (filesystem package
		// not yet migrated). Pass a Nop logger — the slog logger above
		// handles all observability logging.
		w, err := filesystem.NewCertificateWatcher(t.OTLPCertPath, t.OTLPKeyPath, zerolog.Nop(), time.Second*5)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP certificate watcher: %w", err)
		}
		tel.watcher = w
		_, err = w.Start(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to start certificate watcher: %w", err)
		}

		ca, err := os.ReadFile(t.OTLPCAPath)
		if err != nil {
			return nil, fmt.Errorf("can't read ca file from %s", t.OTLPCAPath)
		}
		capool := x509.NewCertPool()
		if !capool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("can't add CA cert to pool")
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

	logger.Info("initializing observability", "OTLPAddr", t.OTLPAddr)

	// --- Metrics ---
	if telOpts.metrics != nil {
		logger.Info("initializing provider", "OTLPAddr", t.OTLPAddr, "type", "metrics")
		interval := t.MetricsCollectionInterval
		if telOpts.metrics.collectionInterval > 0 {
			interval = telOpts.metrics.collectionInterval
		}

		metricOpts := WithDefaultMetricOpts(t.ServiceName)
		metricOpts = append(metricOpts, metric.WithResource(telOpts.resource))
		metricOpts = append(metricOpts, telOpts.metrics.extraOpts...)

		mp, err := InitMetricsProvider(logger, t.OTLPAddr, creds, interval, metricOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metric provider: %w", err)
		}
		tel.meterProvider = mp
		logger.Debug("initialized provider", "OTLPAddr", t.OTLPAddr, "type", "metrics")
	}

	// --- Tracing ---
	if telOpts.tracing != nil {
		logger.Info("initializing provider", "OTLPAddr", t.OTLPAddr, "type", "tracing")

		traceOpts := WithDefaultTracingOpts(t.ServiceName)
		traceOpts = append(traceOpts, sdktrace.WithResource(telOpts.resource))
		traceOpts = append(traceOpts, telOpts.tracing.extraOpts...)

		tp, err := InitTraceProvider(logger, t.ServiceName, t.OTLPAddr, creds, t.Pyroscope, traceOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize trace provider: %w", err)
		}
		tel.tracerProvider = tp
		logger.Debug("initialized provider", "OTLPAddr", t.OTLPAddr, "type", "tracing")
	}

	// --- Profiling ---
	if telOpts.profiling != nil {
		pCfg := telOpts.profiling.pyroscope
		// Fall back to TelemetryConfig values for unset fields.
		if pCfg.URL == "" {
			pCfg.URL = t.Pyroscope.URL
		}
		if pCfg.Token == "" {
			pCfg.Token = t.Pyroscope.Token
		}
		if pCfg.User == "" {
			pCfg.User = t.Pyroscope.User
		}

		if pCfg.URL != "" {
			logger.Info("initializing provider", "url", pCfg.URL, "type", "profiling")

			runtime.SetMutexProfileFraction(5)
			runtime.SetBlockProfileRate(5)

			p, err := pyroscope.Start(pyroscope.Config{
				ApplicationName:   t.ServiceName,
				ServerAddress:     pCfg.URL,
				BasicAuthUser:     pCfg.User,
				BasicAuthPassword: pCfg.Token,
				Logger:            &pyroscopeLogger{logger: logger},
				UploadRate:        time.Second * 60,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to initialize profiling provider: %w", err)
			}
			tel.profiler = p
			logger.Debug("initialized provider", "url", pCfg.URL, "type", "profiling")
		} else {
			logger.Debug("profiling enabled but no URL set, skipping", "type", "profiling")
		}
	}

	// --- Logging (OTLP) ---
	if telOpts.logging != nil {
		logger.Info("initializing provider", "OTLPAddr", t.OTLPAddr, "type", "logging")

		lp, err := InitLogProvider(ctx, logger, t.OTLPAddr, creds, telOpts.resource, telOpts.logging.extraOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize log provider: %w", err)
		}
		tel.logProvider = lp
		logger.Debug("initialized provider", "OTLPAddr", t.OTLPAddr, "type", "logging")
	}

	return tel, nil
}
