package services

import (
	"log/slog"
	"os"
	"testing"

	"github.com/herdkey/hello-go/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

)

func TestEchoService_Echo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	service := services.NewEchoService(logger)

	tests := []struct {
		expected api.EchoMessage
		request  api.EchoMessage
		name     string
	}{
		{
			name: "successful echo",
			request: api.EchoMessage{
				Message: "Hello, World!",
				Author:  "Alice",
			},
			expected: api.EchoMessage{
				Message: "Hello, World!",
				Author:  "Alice",
			},
		},
		{
			name: "empty message and author",
			request: api.EchoMessage{
				Message: "",
				Author:  "",
			},
			expected: api.EchoMessage{
				Message: "",
				Author:  "",
			},
		},
		{
			name: "special characters",
			request: api.EchoMessage{
				Message: "Hello! @#$%^&*()",
				Author:  "User123",
			},
			expected: api.EchoMessage{
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
