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
func (s *EchoService) Echo(msg api.EchoMessage) api.EchoMessage {
	s.logger.Info("Processing echo request",
		"message", msg.Message,
		"author", msg.Author,
	)

	return api.EchoMessage{
		Message: msg.Message,
		Author:  msg.Author,
	}
}
