package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			logger.Info("request completed",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("remote_addr", c.Request().RemoteAddr),
				slog.Int("status", c.Response().Status),
			)
			return err
		}
	}
}
