package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/grafana/otel-profiling-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitTraceProvider(logger zerolog.Logger, serviceName string, addr string, creds *credentials.TransportCredentials, pyroscopeConfig PyroscopeConfig, opts ...sdktrace.TracerProviderOption) (func(), error) {
	var exporter sdktrace.SpanExporter

	if creds != nil {
		logger.Info().Str("addr", addr).Msg("otlp parameters specified. connecting via grpc to addr")
		// If the OpenTelemetry Collector is running on a local cluster (minikube or
		// microk8s), it should be accessible through the NodePort service at the
		// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
		// endpoint of your cluster. If you run the app inside k8s, then you can
		// probably connect directly to the service through dns
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(*creds), grpc.FailOnNonTempDialError(true), grpc.WithBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		// Set up a trace exporter
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	} else {
		f, err := os.CreateTemp("", "tracing")
		logger.Warn().Str("path", f.Name()).Msg("otlp parameters not specified. writing traces to a temporary file. this is NOT recommend in production")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary trace export file: %w", err)
		}

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
	defaultOpts := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithBatcher(exporter)}
	opts = append(defaultOpts, opts...)
	tracerProvider := sdktrace.NewTracerProvider(
		opts...,
	)
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

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to shutdown TracerProvider")
		}
	}, nil
}
