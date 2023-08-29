module github.com/bloominlabs/baseplate-go/config/auth0

go 1.20

require (
	github.com/auth0/go-auth0 v1.0.1
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230822021333-c7741cde2d19
	github.com/hashicorp/go-cleanhttp v0.5.2
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.43.0
)

require (
	github.com/PuerkitoBio/rehttp v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	go.opentelemetry.io/otel v1.17.0 // indirect
	go.opentelemetry.io/otel/metric v1.17.0 // indirect
	go.opentelemetry.io/otel/trace v1.17.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
