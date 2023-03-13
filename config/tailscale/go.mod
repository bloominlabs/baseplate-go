module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230313054937-d8aef9fa22e6
	github.com/tailscale/tailscale-client-go v1.8.0
)

require github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect
