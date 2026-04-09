package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func WithDefaultTracingOpts(serviceName string) []sdktrace.TracerProviderOption {
	res, _ := WithDefaultResource(context.Background(), serviceName)
	return []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}
}

// InitTraceProvider initializes the OTLP trace exporter and returns the
// TracerProvider. If pyroscopeConfig.URL is non-empty, the provider is wrapped
// with the Pyroscope OTel integration. The caller is responsible for calling
// provider.Shutdown.
func InitTraceProvider(logger *slog.Logger, serviceName string, addr string, creds *credentials.TransportCredentials, pyroscopeConfig PyroscopeConfig, opts ...sdktrace.TracerProviderOption) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter

	if creds != nil {
		logger.Info("otlp parameters specified, connecting via grpc", "addr", addr)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(*creds))
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	} else {
		f, err := os.CreateTemp("", "tracing")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary trace export file: %w", err)
		}
		logger.Warn("otlp parameters not specified, writing traces to temporary file", "path", f.Name())

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(f),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	defaultOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	}
	opts = append(defaultOpts, opts...)
	tracerProvider := sdktrace.NewTracerProvider(opts...)

	// Only wrap with Pyroscope when a URL is configured.
	if pyroscopeConfig.URL != "" {
		otel.SetTracerProvider(
			otelpyroscope.NewTracerProvider(tracerProvider,
				otelpyroscope.WithAppName(serviceName),
				otelpyroscope.WithPyroscopeURL(pyroscopeConfig.URL),
				otelpyroscope.WithRootSpanOnly(true),
				otelpyroscope.WithAddSpanName(true),
				otelpyroscope.WithProfileURL(true),
				otelpyroscope.WithProfileBaselineURL(true),
			),
		)
	} else {
		otel.SetTracerProvider(tracerProvider)
	}

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}
