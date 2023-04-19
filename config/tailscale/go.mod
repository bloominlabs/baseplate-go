module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230313054937-d8aef9fa22e6
	github.com/tailscale/tailscale-client-go v1.9.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/tailscale/hujson v0.0.0-20221223112325-20486734a56a // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
