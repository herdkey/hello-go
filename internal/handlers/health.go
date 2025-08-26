package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/herdkey/hello-go/internal/api"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	Logger *slog.Logger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		Logger: logger,
	}
}

// GetHealth handles the /healthz endpoint.
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := api.HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetReady handles the /readyz endpoint.
func (h *HealthHandler) GetReady(w http.ResponseWriter, r *http.Request) {
	response := api.HealthResponse{
		Status: "ready",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Failed to encode readiness response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
