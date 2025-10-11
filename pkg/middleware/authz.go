package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	managementapi "github.com/auth0/go-auth0/management"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/jellydator/ttlcache/v3"
	"github.com/labstack/echo/v4"
	"github.com/tacokumo/admin-api/pkg/a0/management"
)

type authZConfig struct {
	exclusionURLPrefixes []string
}

type AuthZOption func(*authZConfig)

func WithExclusionURLPrefixes(prefixes []string) AuthZOption {
	return func(c *authZConfig) {
		c.exclusionURLPrefixes = prefixes
	}
}

func AuthZ(logger *slog.Logger, managementAPI management.Client, opts ...AuthZOption) echo.MiddlewareFunc {
	cache := ttlcache.New(ttlcache.WithDisableTouchOnHit[string, *managementapi.User]())
	go cache.Start()

	var config authZConfig
	for _, opt := range opts {
		opt(&config)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger.DebugContext(c.Request().Context(), "authz middleware called", slog.String("path", c.Request().URL.Path), slog.String("method", c.Request().Method))

			for _, prefix := range config.exclusionURLPrefixes {
				if strings.HasPrefix(c.Request().URL.Path, prefix) {
					logger.DebugContext(c.Request().Context(), "authz middleware skipped (exclusion)", slog.String("path", c.Request().URL.Path))
					return next(c)
				}
			}

			claims, ok := c.Request().Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			if !ok {
				logger.ErrorContext(c.Request().Context(), "failed to get claims from context", slog.String("path", c.Request().URL.Path), slog.Any("context_value", c.Request().Context().Value(jwtmiddleware.ContextKey{})))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			logger.DebugContext(c.Request().Context(), "token claims", slog.Any("claims", claims))

			sub := claims.RegisteredClaims.Subject

			userItem := cache.Get(sub)
			var user *managementapi.User
			if userItem == nil {
				user, err := managementAPI.GetUser(c.Request().Context(), sub)
				if err != nil {
					logger.Error("failed to get user info from Auth0", slog.String("sub", sub), slog.String("error", err.Error()))
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user info")
				}

				cache.Set(sub, user, 5*time.Minute)
			} else {
				user = userItem.Value()
			}

			permissionList, err := managementAPI.GetUserPermissions(c.Request().Context(), *user.ID)
			if err != nil {
				logger.Error("failed to get user permissions from Auth0", slog.String("user_id", *user.ID), slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user permissions")
			}

			for _, p := range permissionList.Permissions {
				logger.DebugContext(c.Request().Context(), "user permission", slog.String("permission", *p.Name))
			}

			return next(c)
		}
	}
}
