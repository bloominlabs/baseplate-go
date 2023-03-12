module github.com/bloominlabs/hostin-proj/srv-client

go 1.20

require github.com/bloominlabs/baseplate-go/http v0.0.0-00010101000000-000000000000

require (
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/rs/zerolog v1.29.0 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.40.0 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/metric v0.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect
	golang.org/x/sys v0.0.0-20210927094055-39ccf1dd6fa6 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/http => ../../http
