module github.com/bloominlabs/baseplate-go/config/nomad

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230705193734-868eb38c2767
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/nomad/api v0.0.0-20230705142855-ede662a828e1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.48.0
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/cronexpr v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	go.opentelemetry.io/otel v1.23.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.0 // indirect
	go.opentelemetry.io/otel/trace v1.23.0 // indirect
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
