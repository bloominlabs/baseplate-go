module github.com/bloominlabs/baseplate-go/config/filesystem

go 1.20

replace github.com/bloominlabs/baseplate-go/tlsutil => ../../tlsutil/

replace github.com/bloominlabs/baseplate-go/config/env => ../env

require (
	github.com/bloominlabs/baseplate-go/tlsutil v0.0.0-20230313062030-93e37f6e4bfe
	github.com/fsnotify/fsnotify v1.6.0
	github.com/rs/zerolog v1.30.0
	github.com/stretchr/testify v1.8.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
