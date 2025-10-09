package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/savisec/hello-go/internal/api"
)

type ErrorHandler struct {
	Logger *slog.Logger
}

func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{
		Logger: logger,
	}
}

func (eh *ErrorHandler) ServeHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Catch any panic and respond with a 500 error
		defer func() {
			if rec := recover(); rec != nil {
				eh.Logger.Error("Panic recovered", "error", rec)
				writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", eh.Logger)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string, logger *slog.Logger) {
	errorResp := api.Error{
		Error: &errorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		logger.Error("Failed to encode error response", "error", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}
