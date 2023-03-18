package nomad

import (
	"flag"
	"sync"

	"github.com/hashicorp/nomad/api"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type NomadConfig struct {
	sync.RWMutex

	Address string `toml:"address"`
	Token   string `toml:"token"`
	client  *api.Client
}

func (c *NomadConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Address, "nomad.addr", env.GetEnvStrDefault("NOMAD_ADDR", "localhost:4646"), "hostname:port to connect to the nomad server")
	f.StringVar(&c.Token, "noamd.token", env.GetEnvStrDefault("NOMAD_TOKEN", ""), "Token to use to authenticate to nomad")
}

func (c *NomadConfig) Merge(other *NomadConfig) error {
	c.Lock()
	c.Address = other.Address
	c.Token = other.Token
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

func (c *NomadConfig) CreateClient() (*api.Client, error) {
	config := api.DefaultConfig()
	c.RLock()
	config.Address = c.Address
	config.SecretID = c.Token
	c.RUnlock()

	return api.NewClient(config)
}

func (c *NomadConfig) GetClient() (api.Client, error) {
	c.RLock()
	if c.client == nil {
		c.RUnlock()
		client, err := c.CreateClient()
		if err != nil {
			return *client, err
		}
		c.Lock()
		c.client = client
		c.Unlock()

		return *client, err
	}

	defer c.RUnlock()
	return *c.client, nil
}
