package nomad

import (
	"crypto/tls"
	"flag"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/nomad/api"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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

// https://github.com/hashicorp/nomad/blob/fb085186b7874e7d7e008c83e0b443526fb2002e/api/api.go#L273-L286
func defaultHttpClient() *http.Client {
	httpClient := cleanhttp.DefaultPooledClient()
	transport := httpClient.Transport.(*http.Transport)
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Default to http/1: alloc exec/websocket aren't supported in http/2
	// well yet: https://github.com/gorilla/websocket/issues/417
	transport.ForceAttemptHTTP2 = false

	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

	return httpClient
}

func (c *NomadConfig) CreateClient() (*api.Client, error) {
	config := api.DefaultConfig()
	c.RLock()
	config.Address = c.Address
	config.SecretID = c.Token
	c.RUnlock()
	config.HttpClient = defaultHttpClient()

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
		runtime.SetFinalizer(c.client, func(client *api.Client) {
			client.Close()
		})
		c.Lock()
		c.client = client
		c.Unlock()

		return *client, err
	}

	defer c.RUnlock()
	return *c.client, nil
}
