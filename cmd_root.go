package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "google-tasks-to-tasks-org",
		Long: `Convert tasks from Google Tasks to Tasks.org`,
	}
	cmd.AddCommand(convertCmd())
	cmd.AddCommand(diffCmd())
	return cmd
}

func run(ctx context.Context, f func() error) {
	if err := f(); err != nil {
		slog.ErrorContext(ctx, "Error during execution.", "error", err)
		os.Exit(1)
	}
}
