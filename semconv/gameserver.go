package semconv

import "go.opentelemetry.io/otel/attribute"

// The game server in which the application represented by the resource is
// running. The `browser.*` attributes MUST be used only for resources that
// represent applications running relative to a user's game server
const (
	// ServerIDKey is the attribute Key representing the unique ID of a user's server
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: '123456abd', 'Chromium 99', 'Chrome 99'
	// Note: This value is intended to be taken from the environment.
	ServerIDKey = attribute.Key("server.id")

	// The slug for the type of gameserver the game is running
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: 'minecraft-java-edition', 'terraria'
	// Note: This value is intended to be taken from the environment.
	ServerSlugKey = attribute.Key("server.slug")

	// The semver-compatible version the gameserver is currently running
	//
	// Type: string
	// RequirementLevel: Optional
	// Stability: experimental
	// Examples: 'v1.19.3', 'latest'
	// Note: This value is intended to be taken from the environment.
	ServerVersionKey = attribute.Key("server.version")
)

// ServerID returns an attribute KeyValue conforming to the
// "server.id" semantic conventions. It represents the the unique ID of a
// user's server.
func ServerID(val string) attribute.KeyValue {
	return ServerIDKey.String(val)
}
