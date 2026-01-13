package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/tacokumo/admin-api/pkg/auth/session"
)

type SessionContextKey string

const CurrentSessionKey SessionContextKey = "current_session"

var publicPaths = []string{
	"/v1alpha1/health/liveness",
	"/v1alpha1/health/readiness",
	"/v1alpha1/auth/login",
	"/v1alpha1/auth/callback",
}

func SessionMiddleware(
	logger *slog.Logger,
	store session.Store,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			if isPublicPath(path) {
				return next(c)
			}

			sessionID := extractSessionID(c)
			if sessionID == "" {
				logger.DebugContext(c.Request().Context(), "no session id found in request")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication")
			}

			sess, err := store.Get(c.Request().Context(), sessionID)
			if err != nil {
				if errors.Is(err, session.ErrSessionNotFound) {
					logger.DebugContext(c.Request().Context(), "session not found", slog.String("session_id", sessionID))
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid session")
				}
				logger.ErrorContext(c.Request().Context(), "session lookup error", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
			}

			if time.Now().After(sess.ExpiresAt) {
				logger.DebugContext(c.Request().Context(), "session expired", slog.String("session_id", sessionID))
				return echo.NewHTTPError(http.StatusUnauthorized, "session expired")
			}

			ctx := context.WithValue(c.Request().Context(), CurrentSessionKey, sess)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func isPublicPath(path string) bool {
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	return false
}

func extractSessionID(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	cookie, err := c.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func GetCurrentSession(ctx context.Context) *session.Session {
	sess, ok := ctx.Value(CurrentSessionKey).(*session.Session)
	if !ok {
		return nil
	}
	return sess
}
