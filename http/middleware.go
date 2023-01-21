package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// CustomClaims contains custom data we want from the token.
type StratosOAuth2CustomClaims struct {
	Scope  string `json:"scope"`
	UserID string `json:"https://stratos.host/user_id"`
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

// Deprecated: The StratosOAuth2CustomClaims struct now has the 'HasScope'
// function which can be used instead
func HasScope(requestedScope string, scopes []string) bool {
	for _, scope := range scopes {
		if requestedScope == scope {
			return true
		}
	}

	return false
}

// JWTClaimsValue gets the parsed claims from the JWT provided in the request.
// Requires running hte JWTValidatorMiddleware. Can then be used to extra custom claims via
// ```
// validatedClaims, ok := JWTClaimsValue(ctx)
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
func JWTClaimsValue(ctx context.Context) (*validator.ValidatedClaims, bool) {
	raw, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	return raw, ok
}

func CustomClaims(ctx context.Context) (*StratosOAuth2CustomClaims, error) {
	raw, ok := JWTClaimsValue(ctx)
	if !ok {
		return nil, fmt.Errorf("did not find JWT in context with the ContextKey. did you run the JWT middleware?")
	}

	customClaims, ok := raw.CustomClaims.(*StratosOAuth2CustomClaims)
	if !ok {
		return nil, fmt.Errorf("failed to convert custom claims returned to StratosOAuth2CustomClaims")
	}

	return customClaims, nil
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	log.Warn().Stack().Err(err).Msg("user failed to authenticate using JWT")
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

func RatelimiterMiddleware(h http.Handler) http.Handler {
	store, err := memorystore.New(&memorystore.Config{
		// Number of tokens allowed per interval.
		Tokens: 100,

		// Interval until tokens reset.
		Interval: time.Minute,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize store")
	}

	middleware, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
	if err != nil {

		log.Fatal().Err(err).Msg("failed to initialize middleware")
	}

	return middleware.Handle(h)
}

func OTLPHandler(serviceName string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(
			h,
			serviceName,
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
			otelhttp.WithPropagators(propagation.TraceContext{}),
		)
	}
}

func JWTMiddleware(issuerURL string, identifiers []string, opts ...validator.Option) func(http.Handler) http.Handler {
	finalOpts := []validator.Option{
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &StratosOAuth2CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Second * 10),
	}
	finalOpts = append(finalOpts, opts...)
	parsedIssuerURL, err := url.Parse(issuerURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse the issuer url")
	}

	provider := jwks.NewCachingProvider(parsedIssuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		parsedIssuerURL.String(),
		identifiers,
		finalOpts...,
	)
	if err != nil {
		panic(fmt.Errorf("failed to set up the validator: %v", err))
	}
	return jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(ErrorHandler),
	).CheckJWT
}

// RequestHandler adds the trace id as a field to the context's logger
// using fieldKey as field key.
func TraceIDHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			log := zerolog.Ctx(r.Context())
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("traceID", span.SpanContext().TraceID().String())
			})
			next.ServeHTTP(w, r)
		})
	}
}

// Setup
// [github.com/rs/zerolog/hlog](https://github.com/rs/zerolog#integration-with-nethttp)
// integration. Must be called after 'OTLPHandler'.
func HlogHandler(h http.Handler) http.Handler {
	c := alice.New()
	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log.Logger))

	c = c.Append(TraceIDHandler("traceID"))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some prepopulated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
		next.ServeHTTP(w, r)
	})
}
