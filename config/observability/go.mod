module github.com/bloominlabs/baseplate-go/config/observability

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230503052152-c8c9a5e78cd3
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230503052152-c8c9a5e78cd3
	github.com/rs/zerolog v1.29.1
	go.opentelemetry.io/otel v1.15.1
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.38.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.15.1
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.38.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.15.1
	go.opentelemetry.io/otel/metric v0.38.1
	go.opentelemetry.io/otel/sdk v1.15.1
	go.opentelemetry.io/otel/sdk/metric v0.38.1
	google.golang.org/grpc v1.56.0
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.15.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.38.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.15.1 // indirect
	go.opentelemetry.io/otel/trace v1.15.1 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem/

replace github.com/bloominlabs/baseplate-go/config/observability => ../observability
