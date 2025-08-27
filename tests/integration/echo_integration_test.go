package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/herdkey/hello-go/internal/api"
	"github.com/herdkey/hello-go/tests/integration/config"
)

func TestPOSTEcho(t *testing.T) {
	// Load configuration
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	// Create the generated client (point to correct server address).
	client, err := api.NewClientWithResponses(cfg.Server.URL())
	require.NoError(t, err)

	// Prepare a request body matching the EchoMessage struct.
	message := "Hello, Echo!"
	author := "IntegrationTest"
	request := api.EchoMessage{
		Message: message,
		Author:  author,
	}

	// Invoke POST /v1/echo via the generated client.
	resp, err := client.EchoWithResponse(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Confirm the server echoed back the same data.
	require.NotNil(t, resp.JSON200)
	require.Equal(t, message, resp.JSON200.Message)
	require.Equal(t, author, resp.JSON200.Author)
}
