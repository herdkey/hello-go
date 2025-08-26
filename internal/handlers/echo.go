package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/herdkey/hello-go/internal/api"
	"github.com/herdkey/hello-go/internal/services"
)

type EchoHandler struct {
	echoService *services.EchoService
	logger      *slog.Logger
}

func NewEchoHandler(echoService *services.EchoService, logger *slog.Logger) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
		logger:      logger,
	}
}

func (h *EchoHandler) PostV1Echo(w http.ResponseWriter, r *http.Request) {
	var req api.EchoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Message == "" || req.Author == "" {
		h.logger.Error("Missing required fields in request")
		http.Error(w, `{"error":"Missing required fields: message and author"}`, http.StatusBadRequest)
		return
	}

	response := h.echoService.Echo(req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
}
