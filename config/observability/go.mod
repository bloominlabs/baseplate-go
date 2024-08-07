module github.com/bloominlabs/baseplate-go/config/observability

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20231223082533-1ad61761104c
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20231223082533-1ad61761104c
	github.com/bloominlabs/baseplate-go/semconv v0.0.0-20231223082533-1ad61761104c
	github.com/grafana/otel-profiling-go v0.5.1
	github.com/grafana/pyroscope-go v1.1.1
	github.com/rs/zerolog v1.32.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.44.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.24.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/sdk/metric v1.24.0
	google.golang.org/grpc v1.61.1
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/semconv => ../../semconv

replace github.com/bloominlabs/baseplate-go/config/env => ../env

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../filesystem

replace github.com/bloominlabs/baseplate-go/config/observability => ../observability
