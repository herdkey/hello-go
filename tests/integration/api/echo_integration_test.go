package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/savisec/hello-go/internal/api"
	"github.com/savisec/hello-go/tests/integration/config"
)

func TestPOSTEcho(t *testing.T) {
	// Load configuration
	cfg := config.LoadConfig(t)

	// Create the generated client (point to correct server address).
	client, err := api.NewClientWithResponses(cfg.Server.URL())
	require.NoError(t, err)

	// Prepare a request body matching the EchoMessage struct.
	request := api.EchoMessage{
		Message: "Hello, Echo!",
		Author:  "IntegrationTest",
	}

	// Invoke POST /v1/echo via the generated client.
	resp, err := client.EchoWithResponse(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Confirm the server echoed back the same data.
	require.NotNil(t, resp.JSON200)
	require.Equal(t, request.Message, resp.JSON200.Message)
	require.Equal(t, request.Author, resp.JSON200.Author)
}
