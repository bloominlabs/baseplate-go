module github.com/bloominlabs/baseplate-go/config/oidc

go 1.23

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

replace github.com/bloominlabs/baseplate-go/config/observability => ../observability/

replace github.com/bloominlabs/baseplate-go/config/server => ../server/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20241118125315-0b7b36107b4c
	github.com/zitadel/oidc/v3 v3.33.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.57.0
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/zitadel/logging v0.6.1 // indirect
	github.com/zitadel/schema v1.3.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)
