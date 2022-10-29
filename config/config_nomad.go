package config

import (
	"flag"
	"sync"

	"github.com/hashicorp/nomad/api"
)

type NomadConfig struct {
	sync.RWMutex

	Address string
	Token   string
	client  *api.Client
}

func (c *NomadConfig) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&c.Address, "nomad.addr", getenv("NOMAD_ADDR", "localhost:4646"), "hostname:port to connect to the nomad server")
	flag.StringVar(&c.Token, "noamd.token", getenv("NOMAD_TOKEN", ""), "Token to use to authenticate to nomad")
}

func (c *NomadConfig) Merge(other *NomadConfig) error {
	c.Address = other.Address
	c.Token = other.Token

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		c.client = client
	}

	return nil
}

func (c *NomadConfig) CreateClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = c.Address
	config.SecretID = c.Token

	return api.NewClient(config)
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (c *NomadConfig) GetClient() (api.Client, error) {
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
