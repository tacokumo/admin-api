package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
)

type CustomClaims struct {
	Scope       string   `json:"scope"`
	GrantType   string   `json:"gty"`
	Permissions []string `json:"permissions"`
}

// Validate implements validator.CustomClaims.
func (c *CustomClaims) Validate(context.Context) error {
	return nil
}

var _ validator.CustomClaims = (*CustomClaims)(nil)

func JWTMiddleware(
	logger *slog.Logger,
	auth0Domain string,
	cacheTTL time.Duration,
	audience []string) (echo.MiddlewareFunc, error) {
	issuerURL, err := url.Parse("https://" + auth0Domain + "/")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cachingProvider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	jwtValidator, err := validator.New(
		cachingProvider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		audience,
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithExclusionUrls([]string{
			"/v1alpha1/health/liveness",
			"/v1alpha1/health/readiness",
		}),
		jwtmiddleware.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			logger.ErrorContext(r.Context(), "JWT validation error", slog.String("error", err.Error()))
			jwtmiddleware.DefaultErrorHandler(w, r, err)
		}),
	)

	return echo.WrapMiddleware(middleware.CheckJWT), nil
}
