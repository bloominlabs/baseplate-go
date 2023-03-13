module github.com/bloominlabs/baseplate-go/config/nomad

go 1.20

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230313050041-ff362e71dd38
	github.com/hashicorp/nomad/api v0.0.0-20230310211451-9fefc18b7746
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230310171629-522b1b587ee0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bloominlabs/baseplate-go/config/env => ../env/
