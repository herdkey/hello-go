package services_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/herdkey/hello-go/internal/api"
	"github.com/herdkey/hello-go/internal/services"
)

func ptr(s string) *string {
	return &s
}

func TestEchoService_Echo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	service := services.NewEchoService(logger)

	tests := []struct {
		name     string
		request  api.EchoRequest
		expected api.EchoResponse
	}{
		{
			name: "successful echo",
			request: api.EchoRequest{
				Message: ptr("Hello, World!"),
				Author:  ptr("Alice"),
			},
			expected: api.EchoResponse{
				Message: "Hello, World!",
				Author:  "Alice",
			},
		},
		{
			name: "empty message and author",
			request: api.EchoRequest{
				Message: ptr(""),
				Author:  ptr(""),
			},
			expected: api.EchoResponse{
				Message: "",
				Author:  "",
			},
		},
		{
			name: "special characters",
			request: api.EchoRequest{
				Message: ptr("Hello! @#$%^&*()"),
				Author:  ptr("User123"),
			},
			expected: api.EchoResponse{
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

	service := services.NewEchoService(logger)

	require.NotNil(t, service)
	assert.NotNil(t, service.logger)
}
