package consul

import (
	"flag"
	"sync"

	"github.com/hashicorp/consul/api"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type ConsulConfig struct {
	sync.RWMutex

	Address string
	Token   string
	SSL     bool
	client  *api.Client
}

func (c *ConsulConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Address, "consul.addr", env.GetEnvStrDefault("CONSUL_HTTP_ADDR", "localhost:8500"), "hostname:port to connect to the consul server")
	f.StringVar(&c.Token, "consul.token", env.GetEnvStrDefault("CONSUL_HTTP_TOKEN", ""), "Token to use to authenticate to consul")
	f.BoolVar(&c.SSL, "consul.ssl", env.GetEnvBoolDefault("CONSUL_HTTP_SSL", false), "Token to use to authenticate to consul")
}

func (c *ConsulConfig) Merge(other *ConsulConfig) error {
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

func (c *ConsulConfig) CreateClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = c.Address
	config.Token = c.Token

	if c.SSL {
		config.Scheme = "https"
	}

	return api.NewClient(config)
}

func (c *ConsulConfig) GetClient() (api.Client, error) {
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
