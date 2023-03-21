module github.com/bloominlabs/baseplate-go/config/nomad

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230313050041-ff362e71dd38
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/nomad/api v0.0.0-20230310211451-9fefc18b7746
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.40.0
)

require (
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/metric v0.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	golang.org/x/exp v0.0.0-20230310171629-522b1b587ee0 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
