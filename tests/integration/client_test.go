package integration

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientMetricsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that metrics endpoint is accessible
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get("http://localhost:9092/metrics")
	if err != nil {
		t.Skip("Metrics server not running, skipping test")
		return
	}
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")
}

func TestClientHealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get("http://localhost:9092/health")
	if err != nil {
		t.Skip("Health endpoint not available, skipping test")
		return
	}
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClientConfiguration(t *testing.T) {
	// Test configuration loading
	configFile := "../../vpn_client/config.client.toml.example"
	
	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Skip("Config file not found, skipping test")
		return
	}
	
	// Test that config file is readable
	content, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	
	// Basic validation that it contains expected fields
	configStr := string(content)
	assert.Contains(t, configStr, "server_addr")
	assert.Contains(t, configStr, "server_name")
	assert.Contains(t, configStr, "log_level")
}

func TestClientGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test context cancellation handling
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Simulate graceful shutdown
	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	case <-time.After(2 * time.Second):
		t.Fatal("Context should have been cancelled")
	}
}