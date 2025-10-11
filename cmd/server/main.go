package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/tacokumo/admin-api/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	level := os.Getenv("LOG_LEVEL")
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	app := cmd.New(logger)
	if err := app.ExecuteContext(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to execute command", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
