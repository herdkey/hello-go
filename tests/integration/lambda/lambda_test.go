package lambda

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/savisec/hello-go/internal/handlers"
	"github.com/savisec/hello-go/tests/integration/config"
)

// API Gateway V2 HTTP API request format
type LambdaAPIGatewayV2Request struct {
	Headers         map[string]string `json:"headers"`
	Version         string            `json:"version"`
	RouteKey        string            `json:"routeKey"`
	RawPath         string            `json:"rawPath"`
	RawQueryString  string            `json:"rawQueryString"`
	Body            string            `json:"body"`
	RequestContext  RequestContext    `json:"requestContext"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type RequestContext struct {
	AccountID    string `json:"accountId"`
	APIID        string `json:"apiId"`
	DomainName   string `json:"domainName"`
	DomainPrefix string `json:"domainPrefix"`
	HTTP         HTTP   `json:"http"`
	RequestID    string `json:"requestId"`
	RouteKey     string `json:"routeKey"`
	Stage        string `json:"stage"`
	Time         string `json:"time"`
	TimeEpoch    int64  `json:"timeEpoch"`
}

type HTTP struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	SourceIP  string `json:"sourceIp"`
	UserAgent string `json:"userAgent"`
}

type LambdaResponse struct {
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
	Cookies           []string            `json:"cookies"`
	StatusCode        int                 `json:"statusCode"`
}

func TestLambdaHealthEndpoint(t *testing.T) {
	cfg := config.LoadConfig(t)

	lambdaReq := LambdaAPIGatewayV2Request{
		Version:  "2.0",
		RouteKey: "GET /healthz",
		RawPath:  "/healthz",
		Headers: map[string]string{
			"accept": "*/*",
		},
		RequestContext: RequestContext{
			HTTP: HTTP{
				Method:   "GET",
				Path:     "/healthz",
				Protocol: "HTTP/1.1",
				SourceIP: "127.0.0.1",
			},
			RouteKey: "GET /healthz",
			Stage:    "$default",
		},
		IsBase64Encoded: false,
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

	lambdaReq := LambdaAPIGatewayV2Request{
		Version:  "2.0",
		RouteKey: "POST /v1/echo",
		RawPath:  "/v1/echo",
		Headers: map[string]string{
			"accept":       "application/json",
			"content-type": "application/json",
		},
		RequestContext: RequestContext{
			HTTP: HTTP{
				Method:   "POST",
				Path:     "/v1/echo",
				Protocol: "HTTP/1.1",
				SourceIP: "127.0.0.1",
			},
			RouteKey: "POST /v1/echo",
			Stage:    "$default",
		},
		Body:            string(bodyJSON),
		IsBase64Encoded: false,
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

	lambdaReq := LambdaAPIGatewayV2Request{
		Version:  "2.0",
		RouteKey: "POST /v1/echo",
		RawPath:  "/v1/echo",
		Headers: map[string]string{
			"accept":       "application/json",
			"content-type": "application/json",
		},
		RequestContext: RequestContext{
			HTTP: HTTP{
				Method:   "POST",
				Path:     "/v1/echo",
				Protocol: "HTTP/1.1",
				SourceIP: "127.0.0.1",
			},
			RouteKey: "POST /v1/echo",
			Stage:    "$default",
		},
		Body:            string(bodyJSON),
		IsBase64Encoded: false,
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

func invokeLambda(t *testing.T, cfg *config.Config, req LambdaAPIGatewayV2Request) LambdaResponse {
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
