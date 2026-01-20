package earthengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
)

const (
	earthEngineAPIBaseURL = "https://earthengine.googleapis.com/v1"
	earthEngineScope      = "https://www.googleapis.com/auth/earthengine"
)

// Client is the main client for interacting with Google Earth Engine REST API.
type Client struct {
	httpClient *http.Client
	projectID  string
	baseURL    string
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client) error

// NewClient creates a new Earth Engine client with the provided options.
// At minimum, you must provide WithProject and one authentication method.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	client := &Client{
		baseURL: earthEngineAPIBaseURL,
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	if client.projectID == "" {
		return nil, fmt.Errorf("project ID is required (use WithProject option)")
	}

	if client.httpClient == nil {
		return nil, fmt.Errorf("authentication is required (use WithServiceAccountFile, WithServiceAccountJSON, or WithServiceAccountEnv)")
	}

	return client, nil
}

// WithProject sets the Google Cloud project ID.
func WithProject(projectID string) ClientOption {
	return func(c *Client) error {
		if projectID == "" {
			return fmt.Errorf("project ID cannot be empty")
		}
		c.projectID = projectID
		return nil
	}
}

// WithServiceAccountFile sets authentication using a service account JSON key file.
func WithServiceAccountFile(filePath string) ClientOption {
	return func(c *Client) error {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read service account file: %w", err)
		}
		return setupServiceAccountAuth(c, data)
	}
}

// WithServiceAccountJSON sets authentication using service account JSON credentials.
func WithServiceAccountJSON(jsonData []byte) ClientOption {
	return func(c *Client) error {
		return setupServiceAccountAuth(c, jsonData)
	}
}

// WithServiceAccountEnv sets authentication using environment variables.
// Expected environment variables:
// - GOOGLE_EARTH_ENGINE_PROJECT_ID
// - GOOGLE_EARTH_ENGINE_CLIENT_EMAIL
// - GOOGLE_EARTH_ENGINE_PRIVATE_KEY
func WithServiceAccountEnv() ClientOption {
	return func(c *Client) error {
		projectID := os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")
		clientEmail := os.Getenv("GOOGLE_EARTH_ENGINE_CLIENT_EMAIL")
		privateKey := os.Getenv("GOOGLE_EARTH_ENGINE_PRIVATE_KEY")

		if projectID == "" || clientEmail == "" || privateKey == "" {
			return fmt.Errorf("missing required environment variables: GOOGLE_EARTH_ENGINE_PROJECT_ID, GOOGLE_EARTH_ENGINE_CLIENT_EMAIL, GOOGLE_EARTH_ENGINE_PRIVATE_KEY")
		}

		// Construct service account JSON
		saJSON := map[string]string{
			"type":         "service_account",
			"project_id":   projectID,
			"client_email": clientEmail,
			"private_key":  privateKey,
		}

		jsonData, err := json.Marshal(saJSON)
		if err != nil {
			return fmt.Errorf("failed to marshal service account JSON: %w", err)
		}

		return setupServiceAccountAuth(c, jsonData)
	}
}

// setupServiceAccountAuth configures OAuth2 authentication using service account credentials.
func setupServiceAccountAuth(c *Client, jsonData []byte) error {
	ctx := context.Background()
	config, err := google.JWTConfigFromJSON(jsonData, earthEngineScope)
	if err != nil {
		return fmt.Errorf("failed to create JWT config: %w", err)
	}

	c.httpClient = config.Client(ctx)
	return nil
}

// WithHTTPClient sets a custom HTTP client (useful for testing).
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

// ComputeValue executes an Earth Engine expression and returns the computed value.
func (c *Client) ComputeValue(ctx context.Context, expr *Expression) (interface{}, error) {
	url := fmt.Sprintf("%s/projects/%s/value:compute", c.baseURL, c.projectID)

	// Marshal the expression to JSON
	exprJSON, err := json.Marshal(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expression: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(exprJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the result value
	if val, ok := result["result"]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no result in response")
}
