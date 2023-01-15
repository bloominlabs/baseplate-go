package config

import (
	"flag"
	"sync"

	"github.com/tailscale/tailscale-client-go/tailscale"
)

type TailscaleConfig struct {
	sync.RWMutex

	Tailnet  string
	ApiToken string `toml:"api_token"`
	BaseURL  string
	client   *tailscale.Client
}

func (c *TailscaleConfig) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&c.Tailnet, "tailscale.tailnet", GetEnvStrDefault("TAILSCALE_TAILNET", "bloominlabs.com"), "tailnet to perform API operations on")
	flag.StringVar(&c.ApiToken, "tailscale.api-token", GetEnvStrDefault("TAILSCALE_API_TOKEN", ""), "Token to use to authenticate to tailscale")
	flag.StringVar(&c.BaseURL, "tailscale.base-url", GetEnvStrDefault("TAILSCALE_BASE_URL", ""), "Base URL to use for tailscale requests. normally used by tests")
}

func (c *TailscaleConfig) Merge(other *TailscaleConfig) error {
	c.Tailnet = other.Tailnet
	c.ApiToken = other.ApiToken

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.client = client
	}

	return nil
}

func (c *TailscaleConfig) CreateClient() (*tailscale.Client, error) {
	opts := make([]tailscale.ClientOption, 0)
	if c.BaseURL != "" {
		opts = append([]tailscale.ClientOption{tailscale.WithBaseURL(c.BaseURL)})
	}
	return tailscale.NewClient(c.ApiToken, c.Tailnet, opts...)
}

func (c *TailscaleConfig) GetClient() (tailscale.Client, error) {
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
