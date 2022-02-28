package observability

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
)

func WithDefaultResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		// TODO
		// resource.WithAttributes(
		// 	// the service name used to display traces in backends
		// 	semconv.ServiceNameKey.String(config.Service),
		// 	semconv.ServiceVersionKey.String(config.Version),
		// 	attribute.String("environment", config.Environment),
		// 	attribute.Int64("ID", config.ID),
		// ),
	)
}
