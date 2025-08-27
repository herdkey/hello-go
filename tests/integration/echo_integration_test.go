package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/herdkey/hello-go/internal/api"
)

func TestPOSTEcho(t *testing.T) {
	// Create the generated client (point to correct server address).
	client, err := api.NewClientWithResponses("http://localhost:8080")
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
