package integration

import (
    "context"
    "net/http"
    "testing"

    "github.com/stretchr/testify/require"
    // Adjust the import below to match your actual module path.
    "github.com/your-module-name/internal/api"
)

func TestPOSTEcho(t *testing.T) {
    // Create the generated client (point to correct server address).
    client, err := api.NewClientWithResponses("http://localhost:8080")
    require.NoError(t, err)

    // Prepare a request body matching the EchoMessage struct.
    request := api.EchoMessage{
        Message: "Hello, Echo!",
        Author:  "IntegrationTest",
    }

    // Invoke POST /v1/echo via the generated client.
    resp, err := client.PostV1EchoWithResponse(context.Background(), request)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, resp.StatusCode())

    // Confirm the server echoed back the same data.
    require.NotNil(t, resp.JSON200)
    require.Equal(t, request.Message, resp.JSON200.Message)
    require.Equal(t, request.Author, resp.JSON200.Author)
}
