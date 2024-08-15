module github.com/bloominlabs/baseplate-go/config/observability

go 1.21

toolchain go1.22.5

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20240814064914-738b3dabf1ba
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20240814064914-738b3dabf1ba
	github.com/bloominlabs/baseplate-go/semconv v0.0.0-20240814064914-738b3dabf1ba
	github.com/grafana/otel-profiling-go v0.5.1
	github.com/grafana/pyroscope-go v1.1.1
	github.com/rs/zerolog v1.33.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.28.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/sdk/metric v1.28.0
	google.golang.org/grpc v1.65.0
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.7 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.21.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/bloominlabs/baseplate-go/semconv => ../../semconv

replace github.com/bloominlabs/baseplate-go/config/env => ../env

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem

replace github.com/bloominlabs/baseplate-go/config/observability => ../observability
