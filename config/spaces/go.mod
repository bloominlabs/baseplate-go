module github.com/bloominlabs/baseplate-go/config/spaces

go 1.20

replace github.com/bloominlabs/baseplate-go/config/env => ../env/

require (
	github.com/aws/aws-sdk-go-v2 v1.25.2
	github.com/aws/aws-sdk-go-v2/config v1.27.4
	github.com/aws/aws-sdk-go-v2/credentials v1.17.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.50.2
	github.com/bloominlabs/baseplate-go/config/env v0.0.0-20230419034715-89fcb81782b1
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-multierror v1.1.1
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.49.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.27.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.8.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.29.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.1 // indirect
	github.com/aws/smithy-go v1.20.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
)
