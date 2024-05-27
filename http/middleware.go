package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

// CustomClaims contains custom data we want from the token.
type StratosOAuth2CustomClaims struct {
	Scope       string   `json:"scope"`
	UserID      string   `json:"https://stratos.host/user_id"`
	Permissions []string `json:"permissions"`
}

func (c StratosOAuth2CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func (c *StratosOAuth2CustomClaims) HasScope(requestedScope string) bool {
	scopes := strings.Split(c.Scope, " ")
	for _, scope := range scopes {
		if requestedScope == scope {
			return true
		}
	}

	return false
}

func (c *StratosOAuth2CustomClaims) HasPermission(requestedPermission string) bool {
	for _, permission := range c.Permissions {
		if requestedPermission == permission {
			return true
		}
	}

	return false
}

type NomadCustomClaims struct {
	AllocationID string `json:"nomad_allocation_id"`
	JobID        string `json:"nomad_job_id"`
	Namespace    string `json:"nomad_namespace"`
	Task         string `json:"nomad_task"`
	Sub          string `json:"sub"`
}

func (c NomadCustomClaims) Validate(ctx context.Context) error {
	return nil
}

// JWTClaimsValueFromCtx gets the parsed claims from the JWT provided in the request.
// Requires running hte JWTValidatorMiddleware. Can then be used to extra custom claims via
// ```
// validatedClaims, ok := JWTClaimsValueFromCtx(ctx)
//
//	if !ok {
//	 return "", errors.New("failed to decode JWT claims from context")
//	}
//
// claims, ok := validatedClaims.CustomClaims.(*YourCustomClaims)
//
//	if !ok {
//	  return "", errors.New("failed to decode custom claims from the validated claim")
//	}
//
// ```
func JWTClaimsValueFromCtx(ctx context.Context) (*validator.ValidatedClaims, bool) {
	raw, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	return raw, ok
}

func Auth0CustomClaimsFromCtx(ctx context.Context) (*StratosOAuth2CustomClaims, error) {
	raw, ok := JWTClaimsValueFromCtx(ctx)
	if !ok {
		return nil, fmt.Errorf("did not find JWT in context with the ContextKey. did you run the JWT middleware?")
	}

	customClaims, ok := raw.CustomClaims.(*StratosOAuth2CustomClaims)
	if !ok {
		return nil, fmt.Errorf("failed to convert custom claims returned to StratosOAuth2CustomClaims")
	}

	return customClaims, nil
}

func NomadCustomClaimsFromCtx(ctx context.Context) (*NomadCustomClaims, error) {
	raw, ok := JWTClaimsValueFromCtx(ctx)
	if !ok {
		return nil, fmt.Errorf("did not find JWT in context with the ContextKey. did you run the JWT middleware?")
	}

	customClaims, ok := raw.CustomClaims.(*NomadCustomClaims)
	if !ok {
		return nil, fmt.Errorf("failed to convert custom claims returned to StratosOAuth2CustomClaims")
	}

	return customClaims, nil
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	logger := hlog.FromRequest(r)

	logger.Warn().Stack().Err(err).Msg("user failed to authenticate using JWT")
	switch {
	case errors.Is(err, jwtmiddleware.ErrJWTMissing):
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"JWT is missing."}`))
	case errors.Is(err, jwtmiddleware.ErrJWTInvalid):
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"JWT is invalid."}`))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Something went wrong while checking the JWT."}`))
	}
}

type Config struct {
	KeyFunc           httplimit.KeyFunc
	MemoryStoreConfig *memorystore.Config
}

// Option is how options for the JWTMiddleware are set up.
type RatelimiterOption func(*Config)

// WithInterval sets the interval the ratelimiter token bucket refreshes
//
// Default value: time.Second.
func WithInterval(value time.Duration) RatelimiterOption {
	return func(c *Config) {
		c.MemoryStoreConfig.Interval = value
	}
}

func WithTokens(value uint64) RatelimiterOption {
	return func(c *Config) {
		c.MemoryStoreConfig.Tokens = value
	}
}

func WithKeyFunc(value httplimit.KeyFunc) RatelimiterOption {
	return func(c *Config) {
		c.KeyFunc = value
	}
}

func RatelimiterMiddleware(opts ...RatelimiterOption) func(http.Handler) http.Handler {
	config := Config{
		KeyFunc: httplimit.IPKeyFunc(),
		MemoryStoreConfig: &memorystore.Config{
			// Number of tokens allowed per interval.
			Tokens: 10,

			// Interval until tokens reset.
			Interval: time.Second,
		},
	}
	for _, opt := range opts {
		opt(&config)
	}

	store, err := memorystore.New(config.MemoryStoreConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize store")
	}

	middleware, err := httplimit.NewMiddleware(store, config.KeyFunc)
	if err != nil {

		log.Fatal().Err(err).Msg("failed to initialize middleware")
	}

	return middleware.Handle
}

func OTLPHandler(serviceName string, options ...otelhttp.Option) func(http.Handler) http.Handler {
	opts := []otelhttp.Option{
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithPropagators(propagation.TraceContext{}),
	}
	opts = append(opts, options...)

	return func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(
			h,
			serviceName,
			opts...,
		)
	}
}

func jwtMiddleware(issuerURL string, audience []string, customClaims func() validator.CustomClaims, opts ...jwtmiddleware.Option) func(http.Handler) http.Handler {
	client := cleanhttp.DefaultPooledClient()
	client.Transport = otelhttp.NewTransport(client.Transport)

	parsedIssuerURL, err := url.Parse(issuerURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse the issuer url")
	}

	provider := jwks.NewCachingProvider(
		parsedIssuerURL,
		5*time.Minute,
		jwks.WithCustomClient(client),
	)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		parsedIssuerURL.String(),
		audience,
		validator.WithCustomClaims(customClaims),
		validator.WithAllowedClockSkew(time.Second*10),
	)
	if err != nil {
		panic(fmt.Errorf("failed to set up the validator: %v", err))
	}
	finalOpts := []jwtmiddleware.Option{jwtmiddleware.WithErrorHandler(ErrorHandler)}
	finalOpts = append(finalOpts, opts...)
	return jwtmiddleware.New(
		jwtValidator.ValidateToken,
		finalOpts...,
	).CheckJWT
}

func Auth0JWTMiddleware(issuerURL string, audience []string, opts ...jwtmiddleware.Option) func(http.Handler) http.Handler {
	return jwtMiddleware(issuerURL, audience, func() validator.CustomClaims {
		return &StratosOAuth2CustomClaims{}
	}, opts...)
}

func NomadJWTMiddleware(nomadURL string, audience []string, opts ...jwtmiddleware.Option) func(http.Handler) http.Handler {
	return jwtMiddleware(nomadURL, audience, func() validator.CustomClaims {
		return &NomadCustomClaims{}
	}, opts...)
}

// Setup
// [github.com/rs/zerolog/hlog](https://github.com/rs/zerolog#integration-with-nethttp)
// integration. Must be called after 'OTLPHandler'.
func HlogHandler(h http.Handler) http.Handler {
	c := alice.New()
	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log.Logger))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some prepopulated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().Ctx(r.Context()).
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))

	return c.Then(h)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return cors.New(cors.Options{ // Check or reject cors TODO: cors tests
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler(next)
}
