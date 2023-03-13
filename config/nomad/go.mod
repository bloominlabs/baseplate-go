module github.com/bloominlabs/baseplate-go/config/nomad

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-00010101000000-000000000000
	github.com/hashicorp/nomad/api v0.0.0-20230303232206-b07af5761846
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230108222341-4b8118a2686a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
