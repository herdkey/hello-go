package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/herdkey/hello-go/internal/config"
)

// Server wraps the HTTP server and logger.
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

	return r
}

func AddHealthRoutes(r chi.Router) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, e := w.Write([]byte(`{"status":"ok"}`))
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, e := w.Write([]byte(`{"status":"ready"}`))
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
