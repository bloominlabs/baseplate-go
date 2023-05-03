module github.com/bloominlabs/baseplate-go/config/nomad

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230419034715-89fcb81782b1
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/nomad/api v0.0.0-20230421025320-b4e6a70fe69b
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.41.0
)

require (
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	go.opentelemetry.io/otel v1.15.0 // indirect
	go.opentelemetry.io/otel/metric v0.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.15.0 // indirect
	golang.org/x/exp v0.0.0-20230420155640-133eef4313cb // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
