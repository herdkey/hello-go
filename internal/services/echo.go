package services

import (
	"log/slog"
)

type EchoService struct {
	logger *slog.Logger
}

type EchoRequest struct {
	Message string `json:"message"`
	Author  string `json:"author"`
}

type EchoResponse struct {
	Message string `json:"message"`
	Author  string `json:"author"`
}

func NewEchoService(logger *slog.Logger) *EchoService {
	return &EchoService{
		logger: logger,
	}
}

func (s *EchoService) Echo(req EchoRequest) EchoResponse {
	s.logger.Info("Processing echo request", 
		"message", req.Message,
		"author", req.Author,
	)

	return EchoResponse{
		Message: req.Message,
		Author:  req.Author,
	}
}
