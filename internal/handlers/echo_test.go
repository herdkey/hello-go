package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/herdkey/hello-go/internal/services"
)

func TestEchoHandler_PostV1Echo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	echoService := services.NewEchoService(logger)
	handler := NewEchoHandler(echoService, logger)

	tests := []struct {
		requestBody    interface{}
		name           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "successful echo",
			requestBody: services.EchoRequest{
				Message: "Hello, World!",
				Author:  "Alice",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Hello, World!","author":"Alice"}`,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid JSON"}`,
		},
		{
			name: "missing message",
			requestBody: map[string]string{
				"author": "Alice",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Missing required fields: message and author"}`,
		},
		{
			name: "missing author",
			requestBody: map[string]string{
				"message": "Hello, World!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Missing required fields: message and author"}`,
		},
		{
			name: "empty message and author",
			requestBody: map[string]string{
				"message": "",
				"author":  "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Missing required fields: message and author"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer

			if tt.requestBody == "invalid json" {
				body.WriteString("invalid json")
			} else {
				err := json.NewEncoder(&body).Encode(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/echo", &body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.PostV1Echo(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response services.EchoResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				var expectedResponse services.EchoResponse
				err = json.Unmarshal([]byte(tt.expectedBody), &expectedResponse)
				require.NoError(t, err)

				assert.Equal(t, expectedResponse.Message, response.Message)
				assert.Equal(t, expectedResponse.Author, response.Author)
			} else {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
