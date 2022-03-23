package observability

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
)

func WithDefaultResourceOpts() []resource.Option {
	return []resource.Option{
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	}
}

func WithDefaultResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		WithDefaultResourceOpts()...,
	)
}
