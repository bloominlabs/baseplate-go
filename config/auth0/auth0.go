package auth0

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cleanhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bloominlabs/baseplate-go/config/env"
)

const DefaultAuth0Domain = "https://hostin-proj.us.auth0.com"

type Auth0Config struct {
	sync.RWMutex

	Domain       string `toml:"domain"`
	Token        string `toml:"token"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	client       *management.Management
}

func (c *Auth0Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Domain, "auth0.domain", env.GetEnvStrDefault("AUTH0_DOMAIN", DefaultAuth0Domain), "hostname:port to connect to the nomad server")
	f.StringVar(&c.Token, "auth0.token", env.GetEnvStrDefault("AUTH0_TOKEN", ""), "Token to use to authenticate to auth0 (can be used instead of auth0.client_id + auth0.client_secret)")
	f.StringVar(&c.ClientID, "auth0.client_id", env.GetEnvStrDefault("AUTH0_CLIENT_ID", ""), "Auth0 Management Client ID to authenticate to auth0 (can be used instead of auth0.token)")
	f.StringVar(&c.ClientSecret, "auth0.client_secret", env.GetEnvStrDefault("AUTH0_CLIENT_SECRET", ""), "Auth0 Management Client Secret with capability to create users (can be used ins tead of auth0.client_token")
}

func (c *Auth0Config) Validate() error {
	if c.ClientID == "" && c.ClientSecret != "" {
		return fmt.Errorf("'client_secret' is specified, but 'client_id' is empty. Please set 'client_id' or use the auth0 token instead")
	}
	if c.ClientID != "" && c.ClientSecret == "" {
		return fmt.Errorf("'client_id' is specified, but 'client_secret' is empty. Please set 'client_secret' or use the auth0 token instead")
	}
	if c.Token == "" && c.ClientID == "" && c.ClientSecret == "" {
		return fmt.Errorf("'token', 'client_id', nor 'client_secret' is specified. cannot authenticate to auth0 at %s", c.Domain)
	}

	return nil
}

func (c *Auth0Config) Merge(other *Auth0Config) error {
	c.Lock()
	if other.Domain != "" {
		c.Domain = other.Domain
	}
	c.Token = other.Token
	c.ClientID = other.ClientID
	c.ClientSecret = other.ClientSecret
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

func defaultHttpClient() *http.Client {
	httpClient := cleanhttp.DefaultPooledClient()
	transport := httpClient.Transport.(*http.Transport)
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

	return httpClient
}

func (c *Auth0Config) CreateClient() (*management.Management, error) {
	client := defaultHttpClient()
	runtime.SetFinalizer(client, func(client *http.Client) {
		client.CloseIdleConnections()
	})
	c.RLock()
	defer c.RUnlock()
	if c.Token == "" {
		return management.New(
			c.Domain,
			management.WithClientCredentials(context.Background(), c.ClientID, c.ClientSecret),
			management.WithClient(client),
		)
	} else {
		return management.New(c.Domain, management.WithStaticToken(c.Token), management.WithClient(client))
	}
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (c *Auth0Config) GetClient() (management.Management, error) {
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
