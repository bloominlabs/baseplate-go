module github.com/bloominlabs/baseplate-go/config/observability

go 1.25.0

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20260125063911-0aa309a55800
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20260125063911-0aa309a55800
	github.com/bloominlabs/baseplate-go/config/slogger v0.0.0-00010101000000-000000000000
	github.com/bloominlabs/baseplate-go/semconv v0.0.0-20260125063911-0aa309a55800
	github.com/grafana/otel-profiling-go v0.5.1
	github.com/grafana/pyroscope-go v1.2.7
	github.com/rs/zerolog v1.34.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.15.0
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.40.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.40.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.40.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/sdk/log v0.16.0
	go.opentelemetry.io/otel/sdk/metric v1.40.0
	google.golang.org/grpc v1.79.1
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.9 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.40.0 // indirect
	go.opentelemetry.io/otel/log v0.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260223185530-2f722ef697dc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260223185530-2f722ef697dc // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/bloominlabs/baseplate-go/semconv => ../../semconv

replace github.com/bloominlabs/baseplate-go/config/env => ../env

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem

replace github.com/bloominlabs/baseplate-go/config/observability => ../observability

replace github.com/bloominlabs/baseplate-go/config/slogger => ../slogger
