package spaces

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type DigitalOceanSpacesConfig struct {
	sync.RWMutex

	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Region          string `toml:"region"`

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
	f.StringVar(
		&c.Endpoint,
		fmt.Sprintf("%s.endpoint", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ENDPOINT", upperPrefix), "digitaloceanspaces.com"),
		"endpoint to use for requests",
	)
	f.StringVar(
		&c.AccessKeyID,
		fmt.Sprintf("%s.access_key_id", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_ACCESS_KEY_ID", upperPrefix), env.GetEnvStrDefault("AWS_ACCESS_KEY_ID", env.GetEnvStrDefault("SPACES_ACCESS_KEY_ID", ""))),
		"Spaces Access Key ID for authentication",
	)
	f.StringVar(
		&c.SecretAccessKey,
		fmt.Sprintf("%s.secret_access_key", prefix),
		env.GetEnvStrDefault(fmt.Sprintf("%s_SECRET_ACCESS_KEY", upperPrefix), env.GetEnvStrDefault("AWS_SECRET_ACCESS_KEY", env.GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", ""))),
		"Spaces Secret Access Key for authentication",
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
	c.Region = other.Region
	c.Endpoint = other.Endpoint
	c.AccessKeyID = other.AccessKeyID
	c.SecretAccessKey = other.SecretAccessKey
	c.Unlock()

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.client = client
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

	return validationErrors
}

func (c *DigitalOceanSpacesConfig) CreateClient() (*s3.Client, error) {
	// setup s3 sdk for use with digitalocean + opentelemetry
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, awsRegion string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.%s", c.Region, c.Endpoint),
		}, nil
	})
	s3config, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithEndpointResolverWithOptions(resolver),
		awsconfig.WithDefaultRegion(c.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)

	if err != nil {
		return nil, err
	}

	otelaws.AppendMiddlewares(&s3config.APIOptions)

	return s3.NewFromConfig(s3config), nil
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
