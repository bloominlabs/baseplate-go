module github.com/bloominlabs/hostin-proj/loadchecker

go 1.19

require (
	github.com/bloominlabs/baseplate-go/config v0.0.0-20230419034715-89fcb81782b1
	github.com/bloominlabs/baseplate-go/config/observability v0.0.0-20230419034715-89fcb81782b1
	github.com/bloominlabs/baseplate-go/config/server v0.0.0-20230419034715-89fcb81782b1
	github.com/bloominlabs/baseplate-go/http v0.0.0-20230419034715-89fcb81782b1
	github.com/justinas/alice v1.2.0
	github.com/rs/zerolog v1.29.1
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/metric v0.37.0
	go.opentelemetry.io/otel/trace v1.14.0
)

require (
	github.com/auth0/go-jwt-middleware/v2 v2.1.0 // indirect
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230419034715-89fcb81782b1 // indirect
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20230419034715-89fcb81782b1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/pelletier/go-toml/v2 v2.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.40.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.37.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.14.0 // indirect
	go.opentelemetry.io/otel/sdk v1.14.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.37.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/crypto v0.8.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/bloominlabs/baseplate-go => ../../

replace github.com/bloominlabs/baseplate-go/http => ../../http

replace github.com/bloominlabs/baseplate-go/config => ../../config/

replace github.com/bloominlabs/baseplate-go/config/env => ../../config/env/

replace github.com/bloominlabs/baseplate-go/config/server => ../../config/server/

replace github.com/bloominlabs/baseplate-go/config/observability => ../../config/observability/

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../../config/filesystem/
