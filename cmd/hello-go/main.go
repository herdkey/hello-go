package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/herdkey/hello-go/internal/app"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "hello-go",
		Short: "A simple echo HTTP server",
		Long:  "hello-go is a simple HTTP server with echo functionality built with Go",
	}

	rootCmd.AddCommand(newServeCommand())

	return rootCmd
}

func newServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		Long:  "Start the HTTP server with the configured settings",
		RunE:  runServe,
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	application, err := app.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	serverErrCh := make(chan error, 1)

	go func() {
		if err := application.Server.Start(); err != nil {
			serverErrCh <- err
		}
	}()

	select {
	case err := <-serverErrCh:
		if err != nil {
			application.Logger.Error("Server error", "error", err)
			return err
		}
	case <-ctx.Done():
		application.Logger.Info("Shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			application.Logger.Error("Shutdown error", "error", err)
			return err
		}

		application.Logger.Info("Application shutdown complete")
	}

	return nil
}
