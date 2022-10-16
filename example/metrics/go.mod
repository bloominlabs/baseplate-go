module github.com/bloominlabs/hostin-proj/loadchecker

go 1.19

require (
	github.com/bloominlabs/baseplate-go v0.0.0-20221009231558-6268226bb87d
	github.com/rs/zerolog v1.28.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.3
	go.opentelemetry.io/otel v1.11.0
	go.opentelemetry.io/otel/metric v0.32.3
	go.opentelemetry.io/otel/sdk v1.11.0
	go.opentelemetry.io/otel/sdk/metric v0.32.3
	go.opentelemetry.io/otel/trace v1.11.0
	google.golang.org/grpc v1.50.1
)

require (
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.11.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.32.3 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.32.3 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.11.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b // indirect
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
	golang.org/x/text v0.3.8 // indirect
	google.golang.org/genproto v0.0.0-20221014213838-99cd37c6964a // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/bloominlabs/baseplate-go => ../../
