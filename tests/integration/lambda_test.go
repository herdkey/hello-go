package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/herdkey/hello-go/internal/handlers"
	"github.com/herdkey/hello-go/tests/integration/config"
)

type LambdaAPIGatewayRequest struct {
	HTTPMethod     string            `json:"httpMethod"`
	Path           string            `json:"path"`
	Headers        map[string]string `json:"headers"`
	RequestContext RequestContext    `json:"requestContext"`
	Body           string            `json:"body"`
}

type RequestContext struct {
	HTTPMethod string `json:"httpMethod"`
	Path       string `json:"path"`
	Stage      string `json:"stage"`
}

type LambdaResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
}

func TestLambdaHealthEndpoint(t *testing.T) {
	cfg := config.LoadConfig(t)

	lambdaReq := LambdaAPIGatewayRequest{
		HTTPMethod: "GET",
		Path:       "/healthz",
		Headers: map[string]string{
			"Accept": "*/*",
			"Host":   "localhost",
		},
		RequestContext: RequestContext{
			HTTPMethod: "GET",
			Path:       "/healthz",
			Stage:      "prod",
		},
		Body: "",
	}

	response := invokeLambda(t, cfg, lambdaReq)

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	var healthResp handlers.HealthResponse
	if err := json.Unmarshal([]byte(response.Body), &healthResp); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	expectedStatus := "ok"
	if healthResp.Status != expectedStatus {
		t.Errorf("Expected status '%s', got '%s'", expectedStatus, healthResp.Status)
	}
}

func TestLambdaEchoEndpoint(t *testing.T) {
	cfg := config.LoadConfig(t)

	echoBody := map[string]string{
		"message": "Hello Lambda Test",
		"author":  "Integration Test",
	}
	bodyJSON, _ := json.Marshal(echoBody)

	lambdaReq := LambdaAPIGatewayRequest{
		HTTPMethod: "POST",
		Path:       "/v1/echo",
		Headers: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
			"Host":         "localhost",
		},
		RequestContext: RequestContext{
			HTTPMethod: "POST",
			Path:       "/v1/echo",
			Stage:      "prod",
		},
		Body: string(bodyJSON),
	}

	response := invokeLambda(t, cfg, lambdaReq)

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d. Response body: %s", response.StatusCode, response.Body)
	}

	var echoResp map[string]string
	if err := json.Unmarshal([]byte(response.Body), &echoResp); err != nil {
		t.Fatalf("Failed to decode echo response: %v", err)
	}

	if echoResp["message"] != echoBody["message"] {
		t.Errorf("Expected message '%s', got '%s'", echoBody["message"], echoResp["message"])
	}

	if echoResp["author"] != echoBody["author"] {
		t.Errorf("Expected author '%s', got '%s'", echoBody["author"], echoResp["author"])
	}
}

func TestLambdaEchoEndpointValidation(t *testing.T) {
	cfg := config.LoadConfig(t)

	// Test with missing author field
	incompleteBody := map[string]string{
		"message": "Hello without author",
	}
	bodyJSON, _ := json.Marshal(incompleteBody)

	lambdaReq := LambdaAPIGatewayRequest{
		HTTPMethod: "POST",
		Path:       "/v1/echo",
		Headers: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
			"Host":         "localhost",
		},
		RequestContext: RequestContext{
			HTTPMethod: "POST",
			Path:       "/v1/echo",
			Stage:      "prod",
		},
		Body: string(bodyJSON),
	}

	response := invokeLambda(t, cfg, lambdaReq)

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d. Response body: %s", response.StatusCode, response.Body)
	}

	var errorResp map[string]string
	if err := json.Unmarshal([]byte(response.Body), &errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResp["error"] == "" {
		t.Error("Expected error message in response")
	}
}

func invokeLambda(t *testing.T, cfg *config.Config, req LambdaAPIGatewayRequest) LambdaResponse {
	t.Helper()

	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal Lambda request: %v", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(cfg.Lambda.InvocationURL(), "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		t.Fatalf("Failed to invoke Lambda: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	var lambdaResp LambdaResponse
	if err := json.NewDecoder(resp.Body).Decode(&lambdaResp); err != nil {
		t.Fatalf("Failed to decode Lambda response: %v", err)
	}

	return lambdaResp
}
