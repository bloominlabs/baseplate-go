package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog/log"
)

// CustomClaims contains custom data we want from the token.
type StratosOAuth2CustomClaims struct {
	Scope  string `json:"scope"`
	UserID string `json:"https://stratos.host/user_id"`
}

func (c StratosOAuth2CustomClaims) Validate(ctx context.Context) error {
	return nil
}

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
// if !ok {
//  return "", errors.New("failed to decode JWT claims from context")
// }
// claims, ok := validatedClaims.CustomClaims.(*YourCustomClaims)
// if !ok {
//   return "", errors.New("failed to decode custom claims from the validated claim")
// }
// ```
func JWTClaimsValue(ctx context.Context) (*validator.ValidatedClaims, bool) {
	raw, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	return raw, ok
}

func JWTValidatorMiddleware(identifiers []string, issuerURL string, opts ...validator.Option) (*jwtmiddleware.JWTMiddleware, error) {
	parsedIssuerURL, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(parsedIssuerURL, 5*time.Minute)

	opts = append([]validator.Option{validator.WithAllowedClockSkew(time.Second * 10)}, opts...)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		parsedIssuerURL.String(),
		identifiers,
		opts...,
	)
	if err != nil {
		panic(fmt.Errorf("failed to set up the validator: %v", err))
	}
	return jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(ErrorHandler),
	), nil

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
