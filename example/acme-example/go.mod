module github.com/bloominlabs/hostin-proj/acme-example

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config v0.0.0-00010101000000-000000000000
	github.com/bloominlabs/baseplate-go/config/observability v0.0.0-00010101000000-000000000000
	github.com/bloominlabs/baseplate-go/config/server v0.0.0-00010101000000-000000000000
	github.com/bloominlabs/baseplate-go/http v0.0.0-00010101000000-000000000000
	github.com/justinas/alice v1.2.0
	github.com/rs/zerolog v1.29.0
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-00010101000000-000000000000 // indirect
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-00010101000000-000000000000 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grafana/dskit v0.0.0-20230303104220-21930666b68c // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pelletier/go-toml/v2 v2.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.40.0 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.14.0 // indirect
	go.opentelemetry.io/otel/metric v0.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.14.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.29.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/bloominlabs/baseplate-go => ../../

replace github.com/bloominlabs/baseplate-go/http => ../../http

replace github.com/bloominlabs/baseplate-go/config => ../../config/

replace github.com/bloominlabs/baseplate-go/config/env => ../../config/env/

replace github.com/bloominlabs/baseplate-go/config/server => ../../config/server/

replace github.com/bloominlabs/baseplate-go/config/observability => ../../config/observability/

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../../config/filesystem/
