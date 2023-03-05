package config

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type DigitalOceanSpacesConfig struct {
	sync.RWMutex

	Region      string
	Endpoint    string
	AccessKeyID string `toml:"access_key_id"`
	SecretKey   string `toml:"secret_key"`

	client *s3.Client
}

func (c *DigitalOceanSpacesConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Region, "spaces.region", GetEnvStrDefault("SPACES_REGION", "sfo3"), "Region to use in querys")
	f.StringVar(&c.Endpoint, "spaces.endpoint", GetEnvStrDefault("SPACES_ENDPOINT", "digitaloceanspaces.com"), "Endpoint to use for queries")
	f.StringVar(&c.AccessKeyID, "spaces.access_key_id", GetEnvStrDefault("AWS_ACCESS_KEY_ID", GetEnvStrDefault("SPACES_ACCESS_KEY_ID", "")), "Spaces Access Key ID for authentication")
	f.StringVar(&c.SecretKey, "spaces.secret_key", GetEnvStrDefault("AWS_SECRET_ACCESS_KEY", GetEnvStrDefault("SPACES_SECRET_ACCESS_KEY", "")), "Spaces Secret Access Key for authentication")
}

func (c *DigitalOceanSpacesConfig) Merge(other *DigitalOceanSpacesConfig) error {
	c.Region = other.Region
	c.Endpoint = other.Endpoint
	c.AccessKeyID = other.AccessKeyID
	c.SecretKey = other.SecretKey

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.client = client
	}

	return nil
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
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretKey, "")),
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
