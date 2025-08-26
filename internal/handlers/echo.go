package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/herdkey/hello-go/internal/api"
	"github.com/herdkey/hello-go/internal/services"
)

// EchoHandler handles echo-related HTTP requests.
type EchoHandler struct {
	echoService *services.EchoService
	logger      *slog.Logger
}

// NewEchoHandler creates a new EchoHandler with the provided service and logger.
func NewEchoHandler(echoService *services.EchoService, logger *slog.Logger) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
		logger:      logger,
	}
}

// PostV1Echo handles POST requests to the /v1/echo endpoint.
func (h *EchoHandler) PostV1Echo(w http.ResponseWriter, r *http.Request) {
	var req api.EchoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", h.logger)
		return
	}

	if req.Message == "" || req.Author == "" {
		h.logger.Error("Missing required fields in request")
		writeErrorResponse(w, http.StatusBadRequest, "Missing required fields: message and author", h.logger)
		return
	}

	response := h.echoService.Echo(req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", h.logger)
		return
	}
}

// writeErrorResponse sends a structured error response.
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string, logger *slog.Logger) {
	errorResp := api.ErrorResponse{
		Error: &errorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		logger.Error("Failed to encode error response", "error", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}
