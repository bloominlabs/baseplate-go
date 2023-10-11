module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230728211818-6c34eee71023
	github.com/tailscale/tailscale-client-go v1.13.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
