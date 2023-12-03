package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"

	bSemconv "github.com/bloominlabs/baseplate-go/semconv"
)

type nomad struct{}

// Detect returns a *Resource that describes the host being run on.
func (nomad) Detect(ctx context.Context) (*resource.Resource, error) {
	return resource.NewSchemaless(
		bSemconv.NomadJobName(os.Getenv("NOMAD_JOB_NAME")),
		bSemconv.NomadAllocID(os.Getenv("NOMAD_ALLOC_ID")),
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

func WithDefaultResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	opts := WithDefaultResourceOpts()
	opts = append(opts, resource.WithAttributes(
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String(serviceName),
		// attribute.String("environment", config.Environment),
		// attribute.Int64("ID", config.ID),
	))

	return resource.New(ctx,
		opts...,
	)
}
