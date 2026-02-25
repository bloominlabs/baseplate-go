module github.com/bloominlabs/baseplate-go/config/slogger

go 1.22

toolchain go1.22.2

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20240430233630-f1246e02a109
	go.opentelemetry.io/otel/trace v1.26.0
)

require go.opentelemetry.io/otel v1.26.0 // indirect
