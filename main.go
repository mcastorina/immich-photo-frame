package main

import (
	"log/slog"
	"os"

	"immich-photo-frame/internal/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if err := app.Run(); err != nil {
		slog.Error("app failed", "error", err)
		os.Exit(1)
	}
}
