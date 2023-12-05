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
	github.com/rs/zerolog v1.31.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/metric v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
)

require (
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20231203092349-b22b7ae489bc // indirect
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20231203092349-b22b7ae489bc // indirect
	github.com/bloominlabs/baseplate-go/semconv v0.0.0-20231203092349-b22b7ae489bc // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pyroscope-io/client v0.7.2 // indirect
	github.com/pyroscope-io/godeltaprof v0.1.2 // indirect
	github.com/pyroscope-io/otel-profiling-go v0.5.0 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.44.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.21.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)
