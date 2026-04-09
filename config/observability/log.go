package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// InitLogProvider initializes an OTLP log exporter and returns the
// LoggerProvider. The caller is responsible for calling provider.Shutdown.
func InitLogProvider(ctx context.Context, logger *slog.Logger, addr string, creds *credentials.TransportCredentials, res *resource.Resource, opts ...sdklog.LoggerProviderOption) (*sdklog.LoggerProvider, error) {
	var exporterOpts []otlploggrpc.Option

	if creds != nil {
		logger.Info("otlp log parameters specified, connecting via grpc", "addr", addr)

		connCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(*creds))
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection for logs: %w", err)
		}
		_ = connCtx // NewClient doesn't block; context used for timeout scope only
		exporterOpts = append(exporterOpts, otlploggrpc.WithGRPCConn(conn))
	} else {
		logger.Warn("otlp log parameters not specified, using default insecure connection", "addr", addr)
		exporterOpts = append(exporterOpts, otlploggrpc.WithEndpoint(addr))
		exporterOpts = append(exporterOpts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	defaultOpts := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	}
	if res != nil {
		defaultOpts = append(defaultOpts, sdklog.WithResource(res))
	}
	finalOpts := append(defaultOpts, opts...)

	provider := sdklog.NewLoggerProvider(finalOpts...)

	return provider, nil
}

// NewOTLPSlogHandler returns an slog.Handler backed by the otelslog bridge.
// Log records handled by this handler are exported to the OTLP log provider.
func NewOTLPSlogHandler(name string, provider *sdklog.LoggerProvider) slog.Handler {
	return otelslog.NewHandler(name, otelslog.WithLoggerProvider(provider))
}
