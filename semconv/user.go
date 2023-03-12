package semconv

import "go.opentelemetry.io/otel/attribute"

// Describes attributes about a specific startos.host user.
const (
	// UserIDKey is the attribute Key representing the unique ID of a user
	//
	// Type: string[]
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: '123456abd'
	// Note: This value is intended to be taken from the environment.
	UserIDKey = attribute.Key("user.id")

	// UserIDKey is the attribute Key representing the human-readible username of a user
	//
	// Type: string[]
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: 'kschoon'
	// Note: This value is intended to be taken from the environment.
	UsernameKey = attribute.Key("user.name")
)
