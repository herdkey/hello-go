package functional

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/herdkey/hello-go/internal/api"
	"time"

	"github.com/herdkey/hello-go/tests/functional/config"
)


func TestHealthEndpoint(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	healthURL := fmt.Sprintf("%s/healthz", baseURL)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var healthResp api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedStatus := "ok"
	if healthResp.Status != expectedStatus {
		t.Errorf("Expected status '%s', got '%s'", expectedStatus, healthResp.Status)
	}
}
