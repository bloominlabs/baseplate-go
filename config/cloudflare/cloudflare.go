package cloudflare

import (
	"flag"
	"fmt"
	"sync"

	"github.com/cloudflare/cloudflare-go"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type RetryPolicy struct {
	MaxRetries    int `toml:"max_retries"`
	MinRetryDelay int `toml:"min_retry_delay"`
	MaxRetryDelay int `toml:"MaxRetryDelay"`
}

type RatelimitConfiguration struct {
	RequestsPerSecond float64      `toml:"requests_per_second"`
	RetryPolicy       *RetryPolicy `toml:"retry_policy"`
}

type CloudflareConfig struct {
	sync.RWMutex

	Token   string `toml:"api_token"`
	BaseURL string `toml:"base_url"`

	RatelimitConfiguration RatelimitConfiguration `toml:"ratelimit"`

	client *cloudflare.API
}

func (c *CloudflareConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Token, "cloudflare.token", env.GetEnvStrDefault("CLOUDFLARE_API_TOKEN", ""), "Cloudflare API token toauthenticate")
	f.StringVar(&c.BaseURL, "cloudflare.base-url", env.GetEnvStrDefault("CLOUDFLARE_BASE_URL", ""), "Base URL to use for requests. normally used by tests")
}

func (c *CloudflareConfig) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("no cloudflare token provided. please specify 'cloudflare.token' via cli, 'CLOUDFLARE_API_TOKEN' in the environment, or 'api_token' in the config ")
	}
	return nil
}

func (c *CloudflareConfig) Merge(other *CloudflareConfig) error {
	c.Token = other.Token

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.client = client
	}

	return nil
}

func (c *CloudflareConfig) CreateClient() (*cloudflare.API, error) {
	opts := make([]cloudflare.Option, 0)
	if c.BaseURL != "" {
		opts = append(opts, cloudflare.BaseURL(c.BaseURL))
	}

	if c.RatelimitConfiguration.RequestsPerSecond != 0 {
		opts = append(opts, cloudflare.UsingRateLimit(c.RatelimitConfiguration.RequestsPerSecond))
	}

	if c.RatelimitConfiguration.RetryPolicy != nil {
		opts = append(opts, cloudflare.UsingRetryPolicy(c.RatelimitConfiguration.RetryPolicy.MaxRetries, c.RatelimitConfiguration.RetryPolicy.MinRetryDelay, c.RatelimitConfiguration.RetryPolicy.MaxRetryDelay))
	}

	client, err := cloudflare.NewWithAPIToken(c.Token, opts...)
	return client, err
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (c *CloudflareConfig) GetClient() (cloudflare.API, error) {
	if c.client == nil {
		client, err := c.CreateClient()
		if err != nil {
			return cloudflare.API{}, err
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
