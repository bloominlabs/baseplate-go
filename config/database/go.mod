module github.com/bloominlabs/baseplate-go/config/database

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	entgo.io/ent v0.12.4
	github.com/XSAM/otelsql v0.26.0
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230425235927-599945dc67e9
	github.com/go-sql-driver/mysql v1.7.1
	github.com/hashicorp/go-multierror v1.1.1
	go.opentelemetry.io/otel v1.20.0
)

require (
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
)
