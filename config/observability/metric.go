package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func WithDefaultMetricOpts(serviceName string) []metric.Option {
	res, _ := WithDefaultResource(context.Background(), serviceName)
	return []metric.Option{
		metric.WithResource(res),
	}
}

// InitMetricsProvider initializes the OTLP metrics exporter and returns the
// MeterProvider. The caller is responsible for calling provider.Shutdown.
func InitMetricsProvider(logger *slog.Logger, addr string, credentials *credentials.TransportCredentials, collectionInterval time.Duration, opts ...metric.Option) (*metric.MeterProvider, error) {
	var exporter metric.Exporter
	if credentials != nil {
		logger.Info("otlp parameters specified, connecting via grpc", "addr", addr)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		con, err := grpc.NewClient(addr, grpc.WithTransportCredentials(*credentials))
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}
		exporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(con))
		if err != nil {
			return nil, fmt.Errorf("failed to create the collector metric client: %w", err)
		}
	} else {
		f, err := os.CreateTemp("", "metrics")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary metrics export file: %w", err)
		}
		logger.Warn("otlp parameters not specified, writing metrics to temporary file", "path", f.Name())

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")

		exporter, err = stdoutmetric.New(
			stdoutmetric.WithEncoder(enc),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(collectionInterval))
	defaultOpts := []metric.Option{
		metric.WithReader(reader),
	}
	finalOpts := append(defaultOpts, opts...)
	provider := metric.NewMeterProvider(
		finalOpts...,
	)

	otel.SetMeterProvider(provider)

	return provider, nil
}
