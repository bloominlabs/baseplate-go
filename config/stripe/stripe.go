package stripe

import (
	"flag"
	"fmt"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-multierror"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type StripeConfig struct {
	sync.RWMutex

	SecretKey string `toml:"secret_key"`

	client *client.API
}

// set the stripe client manually. useful when writing tests with an already
// initialized client.
//
// WARNING: if you use this parameter, be careful to not use CreateClient() as
// it will overwrite the manually set client. I don't currently have a good
// solution to get around this.
func (c *StripeConfig) WithClient(client *client.API) {
	c.client = client
}

func (c *StripeConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(
		&c.SecretKey,
		"stripe.secret_key",
		env.GetEnvStrDefault("STRIPE_SECRET_KEY", ""),
		"stripe API secret key from the portal",
	)
}

func (c *StripeConfig) Merge(other *StripeConfig) error {
	c.Lock()
	// when the configuration is related, it won't have the defaults from
	// RegisterFlags. This can cause c.Region and c.Endpoint to become empty
	// strings since we rely on the default behavior for those two fields.
	c.SecretKey = other.SecretKey
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

func (c *StripeConfig) Validate() error {
	var validationErrors error
	if c.SecretKey == "" {
		multierror.Append(
			validationErrors,
			fmt.Errorf("no secret key provided. did you specify '-stripe.secret_ke' or 'STRIPE_SECRET_KEY' environment variable?"),
		)
	}

	return validationErrors
}

func (c *StripeConfig) CreateClient() (*client.API, error) {
	c.RLock()

	sc := &client.API{}
	httpClient := cleanhttp.DefaultPooledClient()
	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

	sc.Init(c.SecretKey, stripe.NewBackendsWithConfig(&stripe.BackendConfig{HTTPClient: httpClient}))
	c.RUnlock()

	return sc, nil
}

func (c *StripeConfig) GetClient() (client.API, error) {
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
