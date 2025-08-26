package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	// Add the embed import
	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/herdkey/hello-go/internal/api"

	"github.com/herdkey/hello-go/internal/config"
	"github.com/herdkey/hello-go/internal/handlers"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
}

// New creates a new Server instance with the provided configuration and router.
func New(cfg config.ServerConfig, router chi.Router, logger *slog.Logger) *Server {
	srv := &http.Server{
		Addr:         cfg.Address(),
		Handler:      otelhttp.NewHandler(router, "hello-go"),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		server: srv,
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")

	return s.server.Shutdown(ctx)
}

// NewRouter sets up the router with middlewares and routes.
func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Serve openapi.yaml unconditionally
	r.Get("/api/openapi.yaml", serveOpenAPISpec)

	return r
}

// serveOpenAPISpec handles the /api/openapi.yaml endpoint.
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(api.OpenAPISpec)
	if err != nil {
		// Log the error and send a 500 response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AddHealthRoutes(r chi.Router, logger *slog.Logger) {
	healthHandler := handlers.NewHealthHandler(logger)
	r.Get("/healthz", healthHandler.GetHealth)
	r.Get("/readyz", healthHandler.GetReady)
}
