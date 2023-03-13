module github.com/bloominlabs/baseplate-go/config/filesystem

go 1.20

replace github.com/bloominlabs/baseplate-go/tlsutil => ../../tlsutil/

replace github.com/bloominlabs/baseplate-go/config/env => ../env

require (
	github.com/bloominlabs/baseplate-go/tlsutil v0.0.0-00010101000000-000000000000
	github.com/fsnotify/fsnotify v1.6.0
	github.com/hashicorp/consul/sdk v0.13.1
	github.com/rs/zerolog v1.29.0
	github.com/stretchr/testify v1.8.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220908164124-27713097b956 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
