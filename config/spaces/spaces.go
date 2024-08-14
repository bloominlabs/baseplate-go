package spaces

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"

	"github.com/bloominlabs/baseplate-go/config/env"
)

var isPrefixCompatible *regexp.Regexp = regexp.MustCompile(`^[A-Za-z0-9.]+$`)

// create a (flag comptable, environment variable compatible), respectively,
// prefix. Useful for derived configurations that want to register their own
// flags for custom buckets. DO NOT use when running the 'RegisterFlags'
// function as it will be done on your behalf for the default flags.
func CreatePrefix(prefix string) (string, string) {
	if !isPrefixCompatible.MatchString(prefix) {
		panic(fmt.Sprintf("spaces prefix '%s' must only have alphanumeric characters or periods", prefix))
	}

	return strings.ToLower(prefix), strings.ReplaceAll(strings.ToUpper(prefix), ".", "_")
}

type DigitalOceanSpacesConfig struct {
	sync.RWMutex

	URL             string `toml:"url"`
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Region          string `toml:"region"`
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
	//    	spaces.DigitalOceanSpacesConfig
	//    	ServerJobBucket Bucket `toml:"bucket"`
	//    }
	//    ```
	//
	//    I'm not 100% sure how clean this is, but it atleast gives a good escape
	//    hatch while we figure out the ideal way to do this
	// BucketName  string `toml:"bucket_name"`

	client *s3.Client
}

// set the s3 client manually. useful when writing tests with an already
// initialized client.
//
// WARNING: if you use this parameter, be careful to not use CreateClient() as
// it will overwrite the manually set client. I don't currently have a good
// solution to get around this.
func (c *DigitalOceanSpacesConfig) WithClient(client *s3.Client) {
	c.client = client
}

// Register DigitalOceanSpacesConfig flags with the provided flag.FlagSet. The
// prefix must only container alphanumeric characters or periods for instance -
// `spaces`, `spaces.sfo3`, etc.
func (c *DigitalOceanSpacesConfig) RegisterFlags(f *flag.FlagSet, prefix string) {
	c.prefix = prefix
	prefix, upperPrefix := CreatePrefix(prefix)
	f.StringVar(
		&c.Region,
		fmt.Sprintf("%s.region", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_REGION", upperPrefix), "sfo3"),
		"region to associate the client to",
	)
	f.BoolVar(
		&c.TLSSkipVerify,
		fmt.Sprintf("%s.tls-skip-verify", prefix),
		env.GetEnvBoolDefault(fmt.Sprintf("%s_TLS_SKIP_VERIFY", upperPrefix), false),
		"region to associate the client to",
	)
	f.BoolVar(
		&c.UsePathStyle,
		fmt.Sprintf("%s.use-path-style", prefix),
		env.GetEnvBoolDefault(fmt.Sprintf("%s_USE_PATH_STYLE", upperPrefix), false),
		"use aws path style when making requests",
	)
	f.StringVar(
		&c.Endpoint,
		fmt.Sprintf("%s.endpoint", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ENDPOINT", upperPrefix), "digitaloceanspaces.com"),
		"base endpoint to use for requests. this will be combined with region to form the full URL",
	)
	f.StringVar(
		&c.AccessKeyID,
		fmt.Sprintf("%s.access-key-id", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ACCESS_KEY_ID", upperPrefix), env.GetEnvStrDefault("AWS_ACCESS_KEY_ID", env.GetEnvStrDefault("SPACES_ACCESS_KEY_ID", ""))),
		"Spaces Access Key ID for authentication",
	)
	f.StringVar(
		&c.SecretAccessKey,
		fmt.Sprintf("%s.secret-access-key", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_SECRET_ACCESS_KEY", upperPrefix), env.GetEnvStrDefault("AWS_SECRET_ACCESS_KEY", env.GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", ""))),
		"Spaces Secret Access Key for authentication",
	)
	f.StringVar(
		&c.URL,
		fmt.Sprintf("%s.url", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_URL", upperPrefix), ""),
		"can be used in place of 'region' + 'endpoint' to set the s3 url",
	)
	// see the struct for why this is commented out
	// f.StringVar(
	// 	&c.BucketName,
	// 	fmt.Sprintf("%s.bucket.name", prefix),
	// 	env.GetEnvStrDefault("%S_BUCKET_", env.GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", "")),
	// 	"Spaces Secret Access Key for authentication",
	// )
}

func (c *DigitalOceanSpacesConfig) Merge(other *DigitalOceanSpacesConfig) error {
	c.Lock()
	// when the configuration is related, it won't have the defaults from
	// RegisterFlags. This can cause c.Region and c.Endpoint to become empty
	// strings since we rely on the default behavior for those two fields.
	if other.Region != "" {
		c.Region = other.Region
	}
	if other.Endpoint != "" {
		c.Endpoint = other.Endpoint
	}
	c.AccessKeyID = other.AccessKeyID
	c.SecretAccessKey = other.SecretAccessKey
	c.Unlock()

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.Lock()
		c.client = client
		c.Unlock()
	}

	return nil
}

func (c *DigitalOceanSpacesConfig) Validate() error {
	var validationErrors error
	prefix, upperPrefix := CreatePrefix(c.prefix)
	if c.AccessKeyID == "" {
		multierror.Append(
			validationErrors,
			fmt.Errorf("no access key id provided. did you specify '-%s.access_key_id' or '%s_ACCESS_KEY_ID' environment variable?", prefix, upperPrefix),
		)
	}
	if c.SecretAccessKey == "" {
		multierror.Append(
			validationErrors,
			fmt.Errorf("no secret access key provided. did you specify '-%s.secret_access_key' or '%s_SECRET_ACCESS_KEY' environment variable?", prefix, upperPrefix),
		)
	}

	if c.URL == "" && c.Endpoint == "" {
		multierror.Append(
			validationErrors,
			fmt.Errorf("'url' or 'endpoint' must be specified"),
		)
	}

	return validationErrors
}

func (c *DigitalOceanSpacesConfig) CreateClient() (*s3.Client, error) {
	c.RLock()
	// setup s3 sdk for use with digitalocean + opentelemetry
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, awsRegion string, options ...interface{}) (aws.Endpoint, error) {
		url := fmt.Sprintf("https://%s.%s", c.Region, c.Endpoint)
		if c.URL != "" {
			url = c.URL
		}

		return aws.Endpoint{
			URL: url,
		}, nil
	})

	client := cleanhttp.DefaultPooledClient()
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: c.TLSSkipVerify}
	s3config, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithHTTPClient(client),
		awsconfig.WithEndpointResolverWithOptions(resolver),
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
	}), nil
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (c *DigitalOceanSpacesConfig) GetClient() (s3.Client, error) {
	if c.client == nil {
		client, err := c.CreateClient()
		if err != nil {
			return *client, err
		}
		c.Lock()
		c.client = client
		c.Unlock()

		return *client, err
	}

	c.RLock()
	defer c.RUnlock()
	return *c.client, nil
}
