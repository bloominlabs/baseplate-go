module github.com/bloominlabs/hostin-proj/logging-example

go 1.22

toolchain go1.22.1

require (
	github.com/bloominlabs/baseplate-go/config v0.0.0-20240326235425-6b2c439e5cbc
	github.com/bloominlabs/baseplate-go/config/logger v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.32.0
)

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20240430233630-f1246e02a109 // indirect
	github.com/bloominlabs/baseplate-go/config/filesystem v0.0.0-20240326235425-6b2c439e5cbc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.2.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace github.com/bloominlabs/baseplate-go => ../../

replace github.com/bloominlabs/baseplate-go/http => ../../http

replace github.com/bloominlabs/baseplate-go/config => ../../config/

replace github.com/bloominlabs/baseplate-go/config/env => ../../config/env/

replace github.com/bloominlabs/baseplate-go/config/server => ../../config/server/

replace github.com/bloominlabs/baseplate-go/config/observability => ../../config/observability/

replace github.com/bloominlabs/baseplate-go/config/filesystem => ../../config/filesystem/

replace github.com/bloominlabs/baseplate-go/config/logger => ../../config/logger
