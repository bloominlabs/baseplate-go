module github.com/bloominlabs/baseplate-go/config/cloudflare

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20231114235859-a5f525515384
	github.com/cloudflare/cloudflare-go v0.96.0
)

require (
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.6 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)
