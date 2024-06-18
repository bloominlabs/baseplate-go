module github.com/bloominlabs/baseplate-go/config/auth0

go 1.20

require (
	github.com/auth0/go-auth0 v1.7.0
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230830000604-fc56ee0ccd90
	github.com/hashicorp/go-cleanhttp v0.5.2
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0
)

require (
	github.com/PuerkitoBio/rehttp v1.4.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
