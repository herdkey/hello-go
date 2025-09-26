package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/herdkey/hello-go/internal/handlers"
	"github.com/herdkey/hello-go/internal/httpserver"
	"github.com/herdkey/hello-go/internal/services"
)

// BuildRouter creates and configures the chi router with all routes
func BuildRouter(logger *slog.Logger) chi.Router {
	router := httpserver.NewRouter()

	httpserver.AddHealthRoutes(router, logger)

	echoService := services.NewEchoService(logger)
	echoHandler := handlers.NewEchoHandler(echoService, logger)

	router.Post("/v1/echo", echoHandler.PostV1Echo)

	return router
}
