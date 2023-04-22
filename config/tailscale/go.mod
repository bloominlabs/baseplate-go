module github.com/bloominlabs/baseplate-go/config/tailscale

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-00010101000000-000000000000
	github.com/tailscale/tailscale-client-go v1.9.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/tailscale/hujson v0.0.0-20220506213045-af5ed07155e5 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
