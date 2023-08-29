module github.com/bloominlabs/hostin-proj/loadchecker

go 1.20

replace github.com/bloominlabs/baseplate-go => ../../

replace github.com/bloominlabs/baseplate-go/http => ../../http

replace github.com/bloominlabs/baseplate-go/config => ../../config/

replace github.com/bloominlabs/baseplate-go/config/env => ../../config/env/

replace github.com/bloominlabs/baseplate-go/config/server => ../../config/server/

replace github.com/bloominlabs/baseplate-go/config/observability => ../../config/observability/

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../../config/filesystem/

require (
	github.com/bloominlabs/baseplate-go/config v0.0.0-20230730230838-96a1f424e7f0
	github.com/bloominlabs/baseplate-go/config/observability v0.0.0-20230730230838-96a1f424e7f0
	github.com/bloominlabs/baseplate-go/config/server v0.0.0-20230730230838-96a1f424e7f0
	github.com/bloominlabs/baseplate-go/http v0.0.0-20230730230838-96a1f424e7f0
	github.com/justinas/alice v1.2.0
	github.com/rs/zerolog v1.30.0
	go.opentelemetry.io/otel v1.17.0
	go.opentelemetry.io/otel/metric v1.17.0
	go.opentelemetry.io/otel/trace v1.17.0
)

require (
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230730230838-96a1f424e7f0 // indirect
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230730230838-96a1f424e7f0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pyroscope-io/client v0.7.2 // indirect
	github.com/pyroscope-io/godeltaprof v0.1.2 // indirect
	github.com/pyroscope-io/otel-profiling-go v0.4.0 // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230821184602-ccc8af3d0e93 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230821184602-ccc8af3d0e93 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)
