package earthengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewClient_MissingProject(t *testing.T) {
	ctx := context.Background()
	_, err := NewClient(ctx)
	if err == nil {
		t.Error("Expected error when project is missing")
	}
	if !strings.Contains(err.Error(), "project ID is required") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNewClient_MissingAuth(t *testing.T) {
	ctx := context.Background()
	_, err := NewClient(ctx, WithProject("test-project"))
	if err == nil {
		t.Error("Expected error when authentication is missing")
	}
	if !strings.Contains(err.Error(), "authentication is required") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestWithProject(t *testing.T) {
	client := &Client{}
	opt := WithProject("my-project")
	err := opt(client)
	if err != nil {
		t.Fatalf("WithProject returned error: %v", err)
	}
	if client.projectID != "my-project" {
		t.Errorf("Expected projectID to be 'my-project', got '%s'", client.projectID)
	}
}

func TestWithProject_Empty(t *testing.T) {
	client := &Client{}
	opt := WithProject("")
	err := opt(client)
	if err == nil {
		t.Error("Expected error for empty project ID")
	}
}

// Integration test - only runs if credentials are available
func TestGetTreeCoverage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we have credentials via environment variables
	if os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID") == "" {
		t.Skip("Skipping integration test: GOOGLE_EARTH_ENGINE_PROJECT_ID not set")
	}

	ctx := context.Background()

	// Create client using environment variables
	client, err := NewClient(ctx,
		WithServiceAccountEnv(),
		WithProject(os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test location in Washington state (should have tree coverage)
	latitude := 47.6
	longitude := -120.9

	coverage, err := client.GetTreeCoverage(ctx, latitude, longitude)
	if err != nil {
		t.Fatalf("Failed to get tree coverage: %v", err)
	}

	// Verify the coverage is within valid range
	if coverage < 0 || coverage > 100 {
		t.Errorf("Tree coverage out of range: %f (expected 0-100)", coverage)
	}

	t.Logf("Tree coverage at (%.2f, %.2f): %.2f%%", latitude, longitude, coverage)
}

// Test with mock HTTP client
func TestComputeValue_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/value:compute") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": {"tree_canopy": 42.5}}`))
	}))
	defer server.Close()

	// Create client with mock HTTP client
	client := &Client{
		httpClient: server.Client(),
		projectID:  "test-project",
		baseURL:    server.URL,
	}

	// Create a simple expression
	expr := NewExpression()
	constID := expr.AddConstant(123)
	expr.SetResult(constID)

	// Execute the request
	ctx := context.Background()
	result, err := client.ComputeValue(ctx, expr)
	if err != nil {
		t.Fatalf("ComputeValue failed: %v", err)
	}

	// Verify result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result)
	}

	if val := resultMap["tree_canopy"]; val != 42.5 {
		t.Errorf("Expected tree_canopy to be 42.5, got %v", val)
	}
}

func TestComputeValue_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		projectID:  "test-project",
		baseURL:    server.URL,
	}

	expr := NewExpression()
	constID := expr.AddConstant(123)
	expr.SetResult(constID)

	ctx := context.Background()
	_, err := client.ComputeValue(ctx, expr)
	if err == nil {
		t.Error("Expected error from API, got nil")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestGetTreeCoverage_InvalidCoordinates(t *testing.T) {
	client := &Client{
		httpClient: http.DefaultClient,
		projectID:  "test-project",
	}

	ctx := context.Background()

	// Test invalid latitude
	_, err := client.GetTreeCoverage(ctx, 100, -120)
	if err == nil {
		t.Error("Expected error for invalid latitude")
	}

	// Test invalid longitude
	_, err = client.GetTreeCoverage(ctx, 47, -200)
	if err == nil {
		t.Error("Expected error for invalid longitude")
	}
}
