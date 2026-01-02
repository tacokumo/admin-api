package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type CustomClaims struct {
	ClientID    string   `json:"client_id"`
	Scope       string   `json:"scope"`
	GrantType   string   `json:"gty"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type contextKey string

const ValidatedClaimsKey contextKey = "claims"

func JWTMiddleware(
	logger *slog.Logger,
	region string,
	userPoolID string,
	cacheTTL time.Duration,
	clientIDs []string) (echo.MiddlewareFunc, error) {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: cacheTTL,
	})
	if err != nil {
		logger.Error("failed to get JWKS", slog.String("url", jwksURL), slog.String("error", err.Error()))
		return nil, errors.Wrapf(err, "failed to get JWKS from %s", jwksURL)
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolID)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/v1alpha1/health/liveness" || path == "/v1alpha1/health/readiness" {
				logger.DebugContext(c.Request().Context(), "skipping JWT validation for health check")
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.ErrorContext(c.Request().Context(), "missing authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.ErrorContext(c.Request().Context(), "invalid authorization header format")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, jwks.Keyfunc)
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "JWT validation error", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if !token.Valid {
				logger.ErrorContext(c.Request().Context(), "invalid token")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(*CustomClaims)
			if !ok {
				logger.ErrorContext(c.Request().Context(), "failed to parse claims")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			exp, err := claims.GetExpirationTime()
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "failed to get expiration time", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			if exp != nil && time.Now().After(exp.Time) {
				logger.ErrorContext(c.Request().Context(), "token expired")
				return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
			}

			iss, err := claims.GetIssuer()
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "failed to get issuer", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			if iss != issuer {
				logger.ErrorContext(c.Request().Context(), "invalid issuer", slog.String("issuer", iss))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid issuer")
			}

			// Check client_id claim first (used in client_credentials flow)
			validClientID := false
			if claims.ClientID != "" {
				for _, clientID := range clientIDs {
					if claims.ClientID == clientID {
						validClientID = true
						break
					}
				}
			}

			// If client_id claim is not present or invalid, check aud claim
			if !validClientID {
				aud, err := claims.GetAudience()
				if err != nil {
					logger.ErrorContext(c.Request().Context(), "failed to get audience", slog.String("error", err.Error()))
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
				for _, clientID := range clientIDs {
					for _, a := range aud {
						if a == clientID {
							validClientID = true
							break
						}
					}
					if validClientID {
						break
					}
				}
			}

			if !validClientID {
				logger.ErrorContext(c.Request().Context(), "invalid client ID")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid client ID")
			}

			ctx := context.WithValue(c.Request().Context(), ValidatedClaimsKey, claims)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}, nil
}
