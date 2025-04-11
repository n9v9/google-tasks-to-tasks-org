package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/lmittmann/tint"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	if err := rootCmd().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
