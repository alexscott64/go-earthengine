package apiv1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
)

const (
	// DefaultBasePath is the default base URL for the Earth Engine API.
	DefaultBasePath = "https://earthengine.googleapis.com/"

	// APIVersion is the version of the Earth Engine API this client supports.
	APIVersion = "v1"

	// Scope is the OAuth2 scope required for Earth Engine API access.
	Scope = "https://www.googleapis.com/auth/earthengine"
)

// Service represents the Earth Engine API service client.
// It provides access to all API resources organized by category.
type Service struct {
	client   *http.Client
	BasePath string // Base URL for API requests, default is DefaultBasePath

	// API Resources
	Projects *ProjectsService
}

// NewService creates a new Earth Engine API service client.
//
// It requires authentication via service account credentials.
// Use the With* options to configure authentication.
func NewService(ctx context.Context, opts ...Option) (*Service, error) {
	s := &Service{
		BasePath: DefaultBasePath,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// Validate client is set
	if s.client == nil {
		return nil, fmt.Errorf("no HTTP client configured: use WithServiceAccountFile or WithHTTPClient")
	}

	// Initialize resource services
	s.Projects = &ProjectsService{s: s}
	s.Projects.Value = &ProjectsValueService{s: s}
	s.Projects.Image = &ProjectsImageService{s: s}
	s.Projects.Table = &ProjectsTableService{s: s}
	s.Projects.Operations = &ProjectsOperationsService{s: s}
	s.Projects.Assets = &ProjectsAssetsService{s: s}
	s.Projects.ImageCollection = &ProjectsImageCollectionService{s: s}
	s.Projects.Thumbnails = &ProjectsThumbnailsService{s: s}
	s.Projects.Maps = &ProjectsMapsService{s: s}
	s.Projects.Algorithms = &ProjectsAlgorithmsService{s: s}

	return s, nil
}

// Option is a function that configures a Service.
type Option func(*Service) error

// WithServiceAccountFile configures the service to use a service account JSON key file.
func WithServiceAccountFile(filename string) Option {
	return func(s *Service) error {
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read service account file: %w", err)
		}
		return WithServiceAccountJSON(data)(s)
	}
}

// WithServiceAccountJSON configures the service to use service account JSON credentials.
func WithServiceAccountJSON(jsonData []byte) Option {
	return func(s *Service) error {
		ctx := context.Background()
		config, err := google.JWTConfigFromJSON(jsonData, Scope)
		if err != nil {
			return fmt.Errorf("failed to create JWT config: %w", err)
		}
		s.client = config.Client(ctx)
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client for API requests.
// This is primarily useful for testing.
func WithHTTPClient(client *http.Client) Option {
	return func(s *Service) error {
		s.client = client
		return nil
	}
}

// withBasePath sets a custom base path (primarily for testing).
func withBasePath(basePath string) Option {
	return func(s *Service) error {
		s.BasePath = basePath
		return nil
	}
}

// CallOption represents options for individual API calls.
type CallOption interface {
	Apply(*callOptions)
}

type callOptions struct {
	fields string // Field mask for partial responses
}

// Fields returns a CallOption that specifies which fields to include in the response.
// Use this to reduce response size by requesting only specific fields.
//
// Example: Fields("name,size,bands/name")
func Fields(fields string) CallOption {
	return fieldsOption{fields}
}

type fieldsOption struct{ fields string }

func (o fieldsOption) Apply(opts *callOptions) {
	opts.fields = o.fields
}

// getCallOptions processes CallOptions and returns the configured options.
func getCallOptions(opts []CallOption) *callOptions {
	co := &callOptions{}
	for _, opt := range opts {
		opt.Apply(co)
	}
	return co
}

// makeRequest performs an HTTP request and decodes the JSON response.
func (s *Service) makeRequest(ctx context.Context, method, urlPath string, body interface{}, result interface{}, opts ...CallOption) error {
	// Build URL
	u, err := url.Parse(s.BasePath + urlPath)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters from call options
	co := getCallOptions(opts)
	if co.fields != "" {
		q := u.Query()
		q.Set("fields", co.fields)
		u.RawQuery = q.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyJSON))
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return &apiErr
	}

	// Decode result
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
