package services

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoService_Echo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	
	service := NewEchoService(logger)

	tests := []struct {
		name     string
		request  EchoRequest
		expected EchoResponse
	}{
		{
			name: "successful echo",
			request: EchoRequest{
				Message: "Hello, World!",
				Author:  "Alice",
			},
			expected: EchoResponse{
				Message: "Hello, World!",
				Author:  "Alice",
			},
		},
		{
			name: "empty message and author",
			request: EchoRequest{
				Message: "",
				Author:  "",
			},
			expected: EchoResponse{
				Message: "",
				Author:  "",
			},
		},
		{
			name: "special characters",
			request: EchoRequest{
				Message: "Hello! @#$%^&*()",
				Author:  "User123",
			},
			expected: EchoResponse{
				Message: "Hello! @#$%^&*()",
				Author:  "User123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Echo(tt.request)
			
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.Author, result.Author)
		})
	}
}

func TestNewEchoService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewEchoService(logger)
	
	require.NotNil(t, service)
	assert.NotNil(t, service.logger)
}
