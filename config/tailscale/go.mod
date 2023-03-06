module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-00010101000000-000000000000
	github.com/tailscale/tailscale-client-go v1.8.0
)

require github.com/tailscale/hujson v0.0.0-20220506213045-af5ed07155e5 // indirect
