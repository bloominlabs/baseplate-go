package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/rs/cors"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"

	"github.com/bloominlabs/baseplate-go/config/slogger"
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
// Requires running the JWTValidatorMiddleware. Can then be used to extract custom claims via
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
		return nil, fmt.Errorf("failed to convert custom claims returned to NomadCustomClaims")
	}

	return customClaims, nil
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	logger := slogger.FromContext(r.Context())

	logger.Warn("user failed to authenticate using JWT", "error", err)
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
		slog.Error("failed to initialize store", "error", err)
		os.Exit(1)
	}

	middleware, err := httplimit.NewMiddleware(store, config.KeyFunc)
	if err != nil {
		slog.Error("failed to initialize middleware", "error", err)
		os.Exit(1)
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
		slog.Error("failed to parse the issuer url", "error", err)
		os.Exit(1)
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

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter, supporting http.Flusher etc.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// SlogHandler replaces HlogHandler. It stores an enriched logger (with ip,
// user_agent, referer) in the request context via slogger.NewContext, and
// logs an access line after the request completes.
//
// Must be called after OTLPHandler so that trace context is available.
func SlogHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Enrich the logger with request metadata.
			reqLogger := logger.With(
				"ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
			)

			// Store enriched logger in context for downstream handlers.
			ctx := slogger.NewContext(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			// Wrap the response writer to capture status and size.
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Log access line after request completes.
			reqLogger.Info("",
				"method", r.Method,
				"url", r.URL.String(),
				"status", wrapped.status,
				"size", wrapped.size,
				"duration", time.Since(start),
			)
		})
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return cors.New(cors.Options{ // Check or reject cors TODO: cors tests
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler(next)
}

func CreateWebsocketTransport(v validator.Validator, callback func(context.Context, *validator.ValidatedClaims) (context.Context, error)) transport.Websocket {
	return transport.Websocket{
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			logger := slogger.FromContext(ctx)
			authorization := initPayload.Authorization()
			// HTTP Authorization headers are in the format: <Scheme>[SPACE]<Value>
			// Ref. https://tools.ietf.org/html/rfc7236#section-3
			parts := strings.Split(authorization, " ")

			// Authorization Header is invalid if containing 1 or 0 parts, e.g.:
			// "" || "<Scheme><Value>" || "<Scheme>" || "<Value>"
			if len(parts) > 1 {
				scheme := parts[0]
				// Everything after "<Scheme>" is "<Value>", trimmed
				value := strings.TrimSpace(strings.Join(parts[1:], " "))

				// <Scheme> must be "Bearer"
				if strings.ToLower(scheme) == "bearer" {
					// Since Bearer tokens shouldn't contain spaces (rfc6750#section-2.1)
					// "value" is tokenized, only the first item is used
					token := strings.TrimSpace(strings.Split(value, " ")[0])
					claims, err := v.ValidateToken(ctx, token)
					if err != nil {
						logger.Error("failed to validate websocket token", "error", err)
						return nil, nil, fmt.Errorf("failed to validate JWT. please verify you're logged in")
					}
					validatedClaims, ok := claims.(*validator.ValidatedClaims)
					if !ok {
						logger.Error("failed to validate JWT, could not parse validatedClaims")
						return nil, nil, fmt.Errorf("failed to validate JWT. please verify you're logged in")
					}
					if callback != nil {
						ctx, err = callback(ctx, validatedClaims)
						return ctx, nil, err
					}

					return ctx, nil, nil
				}
			}

			return nil, nil, fmt.Errorf("failed to authenticate you. couldn't find a JWT in the Authorization header using the 'Bearer' scheme")
		},
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
