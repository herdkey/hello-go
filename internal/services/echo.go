package services

import (
	"log/slog"

	"github.com/herdkey/hello-go/internal/api"
)

// EchoService provides echo functionality.
type EchoService struct {
	logger *slog.Logger
}

// NewEchoService creates a new EchoService instance with the provided logger.
func NewEchoService(logger *slog.Logger) *EchoService {
	return &EchoService{
		logger: logger,
	}
}

// Echo processes the EchoRequest and returns an EchoResponse.
func (s *EchoService) Echo(req api.EchoRequest) api.EchoResponse {
	s.logger.Info("Processing echo request",
		"message", req.Message,
		"author", req.Author,
	)

	return api.EchoResponse{
		Message: &req.Message,
		Author:  &req.Author,
	}
}
