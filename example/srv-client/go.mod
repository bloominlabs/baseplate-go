module github.com/bloominlabs/hostin-proj/srv-client

go 1.22

toolchain go1.22.1

require github.com/bloominlabs/baseplate-go/http v0.0.0-20240326235425-6b2c439e5cbc

require (
	github.com/auth0/go-jwt-middleware/v2 v2.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/sethvargo/go-limiter v1.0.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/http => ../../http
