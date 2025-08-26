package services

import (
	"log/slog"

	"github.com/herdkey/hello-go/internal/api"
)

type EchoService struct {
	logger *slog.Logger
}

func NewEchoService(logger *slog.Logger) *EchoService {
	return &EchoService{
		logger: logger,
	}
}

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
