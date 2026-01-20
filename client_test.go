package earthengine

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// Note: Integration tests moved to helpers package (helpers/landcover_test.go)

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

// Note: Coordinate validation tests moved to helpers package
