package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	clientcmd "github.com/tacokumo/admin-api/pkg/client/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var level slog.Level

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	app := clientcmd.New(logger)
	if err := app.ExecuteContext(ctx); err != nil {
		logger.ErrorContext(ctx, "application error", "error", err)
		os.Exit(1)
	}
}
