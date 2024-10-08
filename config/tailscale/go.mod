module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.22.0

toolchain go1.22.7

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230728211818-6c34eee71023
	github.com/tailscale/tailscale-client-go v1.17.0
)

require (
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect
	golang.org/x/oauth2 v0.19.0 // indirect
)
