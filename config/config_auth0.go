package config

import (
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/auth0/go-auth0/management"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const DefaultAuth0Domain = "https://hostin-proj.us.auth0.com"

type Auth0Config struct {
	sync.RWMutex

	Domain       string
	Token        string
	ClientID     string
	ClientSecret string
	client       *management.Management
}

func (c *Auth0Config) RegisterFlags(f *flag.FlagSet) {
	flag.StringVar(&c.Domain, "auth0.domain", GetEnvStrDefault("AUTH0_DOMAIN", DefaultAuth0Domain), "hostname:port to connect to the nomad server")
	flag.StringVar(&c.Token, "auth0.token", GetEnvStrDefault("AUTH0_TOKEN", ""), "Token to use to authenticate to auth0 (can be used instead of auth0.client_id + auth0.client_secret)")
	flag.StringVar(&c.ClientID, "auth0.client_id", GetEnvStrDefault("AUTH0_CLIENT_ID", ""), "Auth0 Management Client ID to authenticate to auth0 (can be used instead of auth0.token)")
	flag.StringVar(&c.ClientSecret, "auth0.client_secret", GetEnvStrDefault("AUTH0_CLIENT_SECRET", ""), "Auth0 Management Client Secret with capability to create users (can be used ins tead of auth0.client_token")
}

func (c *Auth0Config) Merge(other *Auth0Config) error {
	c.Domain = other.Domain
	c.Token = other.Token
	c.ClientID = other.ClientID
	c.ClientSecret = other.ClientSecret

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

func (c *Auth0Config) CreateClient() (*management.Management, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	client := &http.Client{Transport: otelhttp.NewTransport(transport), Timeout: time.Minute}

	if c.Token == "" {
		return management.New(c.Domain, management.WithClientCredentials(c.ClientID, c.ClientSecret), management.WithClient(client))
	} else {
		return management.New(c.Domain, management.WithStaticToken(c.Token), management.WithClient(client))
	}
}

// Initialize Metrics + Tracing for the app. NOTE: you must call defer t.Stop() to propely cleanup
func (c *Auth0Config) GetClient() (management.Management, error) {
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
