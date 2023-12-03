package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

type nomad struct{}

// Detect returns a *Resource that describes the host being run on.
func (nomad) Detect(ctx context.Context) (*resource.Resource, error) {
	return resource.NewSchemaless(
		attribute.String("nomad.job.name", os.Getenv("NOMAD_JOB_NAME")),
		attribute.String("nomad.alloc.id", os.Getenv("NOMAD_ALLOC_ID")),
	), nil
}

// WithHost adds attributes from the host to the configured resource.
func WithNomad() resource.Option {
	return resource.WithDetectors(nomad{})
}

func WithDefaultResourceOpts() []resource.Option {
	return []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		WithNomad(),
	}
}

func WithDefaultResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		WithDefaultResourceOpts()...,
	)
}
