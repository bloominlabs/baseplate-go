package s3

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	// "github.com/aws/smithy-go/metrics/smithyotelmetrics"
	// "github.com/aws/smithy-go/tracing/smithyoteltracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	// "go.opentelemetry.io/otel"

	"github.com/bloominlabs/baseplate-go/config/env"
)

var isPrefixCompatible *regexp.Regexp = regexp.MustCompile(`^[A-Za-z0-9.]+$`)

// CreatePrefix create a (flag comptable, environment variable compatible), respectively,
// prefix. Useful for derived configurations that want to register their own
// flags for custom buckets. DO NOT use when running the 'RegisterFlags'
// function as it will be done on your behalf for the default flags.
func CreatePrefix(prefix string) (string, string) {
	if !isPrefixCompatible.MatchString(prefix) {
		panic(fmt.Sprintf("spaces prefix '%s' must only have alphanumeric characters or periods", prefix))
	}

	return strings.ToLower(prefix), strings.ReplaceAll(strings.ToUpper(prefix), ".", "_")
}

type S3Config struct {
	sync.RWMutex

	Endpoint        string `toml:"endpoint"`
	Region          string `toml:"region"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	UsePathStyle    bool   `toml:"use_path_style"`
	TLSSkipVerify   bool   `toml:"tls_skip_verify"`

	prefix string

	// TODO: figure out if we want to handle buckets in the baseplate-go
	// configuration. Counter examples could be:
	// 1. a user may want multiple buckets associated with a single 'client'.
	//    This is not something baseplate-go could support nicely, but could be done
	//    easily with Composition i.e.
	//
	//    ```go
	//    type Bucket struct {
	//    	Name   string `toml:"name"`
	//    }
	//
	//    type SpacesConfig struct {
	//    	spaces.S3Config
	//    	ServerJobBucket Bucket `toml:"bucket"`
	//    }
	//    ```
	//
	//    I'm not 100% sure how clean this is, but it atleast gives a good escape
	//    hatch while we figure out the ideal way to do this
	// BucketName  string `toml:"bucket_name"`

	client          *s3.Client
	transferManager *transfermanager.Client
}

// WithClient sets the s3 client manually. useful when writing tests with an already
// initialized client.
//
// WARNING: if you use this parameter, be careful to not use CreateClient() as
// it will overwrite the manually set client. I don't currently have a good
// solution to get around this.
func (c *S3Config) WithClient(client *s3.Client) {
	c.client = client
}

func (c *S3Config) WithTransferManager(tm *transfermanager.Client) {
	c.transferManager = tm
}

type Option func(*S3Config)

func WithPrefix(prefix string) Option {
	return func(c *S3Config) {
		c.prefix = prefix
	}
}

// RegisterFlags registers S3Config flags with the provided flag.FlagSet. The
// prefix must only container alphanumeric characters or periods for instance -
// `spaces`, `spaces.sfo3`, etc.
func (c *S3Config) RegisterFlags(f *flag.FlagSet, opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.prefix == "" {
		c.prefix = "s3"
	}
	prefix, upperPrefix := CreatePrefix(c.prefix)
	f.BoolVar(
		&c.TLSSkipVerify,
		fmt.Sprintf("%s.tls-skip-verify", prefix),
		env.GetEnvBoolDefault(fmt.Sprintf("%s_TLS_SKIP_VERIFY", upperPrefix), false),
		"should the client skip TLS verification when making requests",
	)
	f.BoolVar(
		&c.UsePathStyle,
		fmt.Sprintf("%s.use-path-style", prefix),
		env.GetEnvBoolDefault(fmt.Sprintf("%s_USE_PATH_STYLE", upperPrefix), false),
		"use aws path style when making requests",
	)
	f.StringVar(
		&c.Region,
		fmt.Sprintf("%s.region", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_REGION", upperPrefix), "us-east-1"),
		"region to associate the client to",
	)
	f.StringVar(
		&c.Endpoint,
		fmt.Sprintf("%s.endpoint", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ENDPOINT", upperPrefix), ""),
		"base endpoint to use for requests",
	)
	f.StringVar(
		&c.AccessKeyID,
		fmt.Sprintf("%s.access-key-id", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ACCESS_KEY_ID", upperPrefix), env.GetEnvStrDefault("AWS_ACCESS_KEY_ID", env.GetEnvStrDefault("SPACES_ACCESS_KEY_ID", ""))),
		"S3 Access Key ID for authentication",
	)
	f.StringVar(
		&c.SecretAccessKey,
		fmt.Sprintf("%s.secret-access-key", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_SECRET_ACCESS_KEY", upperPrefix), env.GetEnvStrDefault("AWS_SECRET_ACCESS_KEY", env.GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", ""))),
		"S3 Secret Access Key for authentication",
	)
	// see the struct for why this is commented out
	// f.StringVar(
	// 	&c.BucketName,
	// 	fmt.Sprintf("%s.bucket.name", prefix),
	// 	env.GetEnvStrDefault("%S_BUCKET_", env.GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", "")),
	// 	"Spaces Secret Access Key for authentication",
	// )
}

func (c *S3Config) Merge(other *S3Config) error {
	c.Lock()
	// when the configuration is related, it won't have the defaults from
	// RegisterFlags. This can cause c.Region and c.Endpoint to become empty
	// strings since we rely on the default behavior for those two fields.
	c.Endpoint = other.Endpoint
	c.AccessKeyID = other.AccessKeyID
	c.SecretAccessKey = other.SecretAccessKey
	c.Unlock()

	client, err := c.CreateClient()
	if err != nil {
		return fmt.Errorf("failed to create s3 client: %w", err)
	}

	tm, err := c.CreateTransferManager()
	if err != nil {
		return fmt.Errorf("failed to create transfer manager: %w", err)
	}

	c.Lock()
	c.client = client
	c.transferManager = tm
	c.Unlock()

	return nil
}

func (c *S3Config) Validate() error {
	var validationErrors error
	prefix, upperPrefix := CreatePrefix(c.prefix)
	if c.AccessKeyID == "" {
		validationErrors = errors.Join(
			validationErrors,
			fmt.Errorf("no access key id provided. did you specify '-%s.access_key_id' or '%s_ACCESS_KEY_ID' environment variable?", prefix, upperPrefix),
		)
	}
	if c.SecretAccessKey == "" {
		validationErrors = errors.Join(
			validationErrors,
			fmt.Errorf("no secret access key provided. did you specify '-%s.secret_access_key' or '%s_SECRET_ACCESS_KEY' environment variable?", prefix, upperPrefix),
		)
	}

	return validationErrors
}

func (c *S3Config) CreateClient() (*s3.Client, error) {
	c.RLock()
	s3config, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(c.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)
	c.RUnlock()
	if err != nil {
		return nil, err
	}

	otelaws.AppendMiddlewares(&s3config.APIOptions)

	return s3.NewFromConfig(s3config, func(o *s3.Options) {
		o.UsePathStyle = c.UsePathStyle
		o.BaseEndpoint = aws.String(c.Endpoint)
		// https://github.com/aws/aws-sdk-go-v2/discussions/2810
		// o.MeterProvider = smithyotelmetrics.Adapt(otel.GetMeterProvider())
		// o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	}), nil
}

func (c *S3Config) GetClient() (*s3.Client, error) {
	if c.client == nil {
		client, err := c.CreateClient()
		if err != nil {
			return client, err
		}
		c.Lock()
		c.client = client
		c.Unlock()

		return client, err
	}

	c.RLock()
	defer c.RUnlock()
	return c.client, nil
}

func (c *S3Config) CreateTransferManager() (*transfermanager.Client, error) {
	client, err := c.GetClient()
	if err != nil {
		return nil, fmt.Errorf("could not create s3 client: %w", err)
	}
	c.Lock()
	c.transferManager = transfermanager.New(client)
	c.Unlock()

	return c.transferManager, nil
}

func (c *S3Config) GetTransferManager() (*transfermanager.Client, error) {
	if c.transferManager == nil {
		tm, err := c.CreateTransferManager()
		if err != nil {
			return nil, fmt.Errorf("could not create transfer manager: %w", err)
		}
		c.Lock()
		c.transferManager = tm
		c.Unlock()

		return tm, nil
	}

	c.RLock()
	defer c.RUnlock()
	return c.transferManager, nil
}
