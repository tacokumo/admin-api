package middleware

import (
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(auth0Domain string, cacheTTL time.Duration, audience []string) (echo.MiddlewareFunc, error) {
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
	)

	return echo.WrapMiddleware(middleware.CheckJWT), nil
}
