package semconv

import "go.opentelemetry.io/otel/attribute"

// Describes attributes about a specific startos.host user.
const (
	// NomadAllocIDKey is the attribute Key representing the allocation ID of the task
	// Additional documentation https://developer.hashicorp.com/nomad/docs/runtime/environment
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: '123456abd'
	// Note: This value is intended to be taken from the environment.
	NomadAllocIDKey = attribute.Key("nomad.alloc.id")

	// NomadJobNameKey is the attribute Key representing the human-readible nomad job name.
	// Additional documentation https://developer.hashicorp.com/nomad/docs/runtime/environment
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: 'kschoon'
	// Note: This value is intended to be taken from the environment.
	NomadJobNameKey = attribute.Key("nomad.job.name")
)

// NomadAllocID returns an attribute KeyValue conforming to the
// "nomad.alloc.id" semantic conventions. It represents the allocation ID of
// the nomad task.
func NomadAllocID(val string) attribute.KeyValue {
	return NomadAllocIDKey.String(val)
}

// NomadJobNAme returns an attribute KeyValue conforming to the
// "nomad.job.name" semantic conventions. It represents the human-readible
// nomad job name.
func NomadJobName(val string) attribute.KeyValue {
	return NomadJobNameKey.String(val)
}
