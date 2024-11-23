package oidc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bloominlabs/baseplate-go/config/env"
)

type ContextKey struct{}

type OIDCConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	Issuer       string `toml:"issuer"`

	client *rs.ResourceServer
	sync.RWMutex
}

func (c *OIDCConfig) Validate() error {
	var validationErrors error
	if c.ClientID == "" {
		validationErrors = errors.Join(validationErrors, fmt.Errorf("failed to validate 'client_id'. did you specify 'oidc.client-id' or 'OIDC_CLIENT_ID'?"))
	}
	if c.ClientSecret == "" {
		validationErrors = errors.Join(validationErrors, fmt.Errorf("failed to validate 'client_secret'. did you specify 'oidc.client-secret' or 'OIDC_CLIENT_SECRET'?"))
	}
	if c.Issuer == "" {
		validationErrors = errors.Join(validationErrors, fmt.Errorf("failed to validate 'issuer'. did you specify 'oidc.issuer' or 'OIDC_ISSUER'?"))
	}
	return validationErrors
}

func (c *OIDCConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.ClientID, "oidc.client-id", env.GetEnvStrDefault("OIDC_CLIENT_ID", ""), "client id to use to authenticate to the issuer")
	f.StringVar(&c.ClientSecret, "oidc.client-secret", env.GetEnvStrDefault("OIDC_CLIENT_SECRET", ""), "client secret to use to authenticate to the issuer")
	f.StringVar(&c.Issuer, "oidc.issuer", env.GetEnvStrDefault("OIDC_ISSUER", ""), "Issuer URL to use for oidc authentication")
}

func (c *OIDCConfig) Merge(o *OIDCConfig) error {
	return nil
}

func (c *OIDCConfig) CreateClient() (*rs.ResourceServer, error) {
	rs, err := rs.NewResourceServerClientCredentials(
		context.Background(),
		c.Issuer,
		c.ClientID,
		c.ClientSecret,
		rs.WithClient(&http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport), Timeout: time.Second * 5}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource server: %w", err)
	}

	return &rs, err
}

func (c *OIDCConfig) GetClient() (*rs.ResourceServer, error) {
	if c.client == nil {
		client, err := c.CreateClient()
		if err != nil {
			return client, err
		}
		c.Lock()
		c.client = client
		c.Unlock()

		return client, err
	}

	c.RLock()
	defer c.RUnlock()
	return c.client, nil
}

func IntrospectionRequestFromContext(ctx context.Context) (*oidc.IntrospectionResponse, bool) {
	raw, ok := ctx.Value(ContextKey{}).(*oidc.IntrospectionResponse)
	return raw, ok
}

func OIDCAuthenticationMiddelware(cfg *OIDCConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header is missing", http.StatusUnauthorized)
				return
			}

			prefix := "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "authorization header is malformed", http.StatusUnauthorized)
				return
			}

			accessToken := strings.TrimPrefix(authHeader, prefix)

			rsClient, err := cfg.GetClient()
			if err != nil {
				http.Error(w, fmt.Errorf("failed to get introspection endpoint: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			resp, err := rs.Introspect[*oidc.IntrospectionResponse](r.Context(), *rsClient, accessToken)
			if err != nil {
				http.Error(w, fmt.Errorf("failed to introspect: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			if !resp.Active {
				http.Error(w, fmt.Errorf("invalid or inactive token").Error(), http.StatusUnauthorized)
			}

			r = r.Clone(context.WithValue(r.Context(), ContextKey{}, resp))
			next.ServeHTTP(w, r)
		})
	}
}
