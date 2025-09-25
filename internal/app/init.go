package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/herdkey/hello-go/internal/config"
	"github.com/herdkey/hello-go/internal/httpserver"
	"github.com/herdkey/hello-go/internal/logging"
	"github.com/herdkey/hello-go/internal/router"
	"github.com/herdkey/hello-go/internal/telemetry"
)

// Application encompasses the server, telemetry, configuration, and logger.
type Application struct {
	Server            *httpserver.Server
	TelemetryProvider *telemetry.Provider
	Config            *config.Config
	Logger            *slog.Logger
}

// Initialize sets up the application with server, telemetry, config, and logging.
func Initialize(ctx context.Context) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger := logging.Setup(cfg.Logging)

	telemetryProvider, err := telemetry.Setup(ctx, cfg.Telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to setup telemetry: %w", err)
	}

	r := router.BuildRouter(logger)

	server := httpserver.New(cfg.Server, r, logger)

	return &Application{
		Server:            server,
		TelemetryProvider: telemetryProvider,
		Config:            cfg,
		Logger:            logger,
	}, nil
}

func (app *Application) Shutdown(ctx context.Context) error {
	app.Logger.Info("Shutting down application")

	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	if err := app.TelemetryProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown telemetry: %w", err)
	}

	return nil
}
