package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	// semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func WithDefaultTracingOpts() []sdktrace.TracerProviderOption {
	res, _ := WithDefaultResource(context.Background())
	return []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}
}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
//
func InitTraceProvider(addr string, creds *credentials.TransportCredentials, opts ...sdktrace.TracerProviderOption) (func(), error) {
	ctx := context.Background()

	grpcCreds := insecure.NewCredentials()
	if creds != nil {
		grpcCreds = *creds
	}

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(grpcCreds), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	defaultOpts := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithSpanProcessor(bsp)}
	opts = append(defaultOpts, opts...)
	tracerProvider := sdktrace.NewTracerProvider(
		opts...,
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		log.Fatal().Err(tracerProvider.Shutdown(ctx)).Msg("failed to shutdown TracerProvider")
	}, nil
}
