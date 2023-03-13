package semconv

import "go.opentelemetry.io/otel/attribute"

// Describes attributes about a specific startos.host user.
const (
	// UserIDKey is the attribute Key representing the unique ID of a user
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: '123456abd'
	// Note: This value is intended to be taken from the environment.
	UserIDKey = attribute.Key("user.id")

	// UsernameKey is the attribute Key representing the human-readible username
	// of a user
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: 'kschoon'
	// Note: This value is intended to be taken from the environment.
	UsernameKey = attribute.Key("user.name")
)

// UserID returns an attribute KeyValue conforming to the
// "user.id" semantic conventions. It represents the the unique ID of a
// user.
func UserID(val string) attribute.KeyValue {
	return UserIDKey.String(val)
}

// Username returns an attribute KeyValue conforming to the
// "user.name" semantic conventions. It represents the human-readible name of
// user.
func Username(val string) attribute.KeyValue {
	return UsernameKey.String(val)
}
