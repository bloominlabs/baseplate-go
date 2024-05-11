module github.com/bloominlabs/baseplate-go/config/logger

go 1.22

toolchain go1.22.2

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20240430233630-f1246e02a109
	github.com/rs/zerolog v1.32.0
	go.opentelemetry.io/otel/trace v1.26.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)
