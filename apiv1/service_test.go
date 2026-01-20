package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewService(t *testing.T) {
	ctx := context.Background()

	t.Run("no auth fails", func(t *testing.T) {
		_, err := NewService(ctx)
		if err == nil {
			t.Error("Expected error when no auth provided")
		}
	})

	t.Run("with http client succeeds", func(t *testing.T) {
		service, err := NewService(ctx, WithHTTPClient(http.DefaultClient))
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if service.BasePath != DefaultBasePath {
			t.Errorf("Expected BasePath %s, got %s", DefaultBasePath, service.BasePath)
		}
		if service.Projects == nil {
			t.Error("Projects service not initialized")
		}
	})

	t.Run("custom base path", func(t *testing.T) {
		customPath := "https://custom.example.com/"
		service, err := NewService(ctx,
			WithHTTPClient(http.DefaultClient),
			withBasePath(customPath),
		)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if service.BasePath != customPath {
			t.Errorf("Expected BasePath %s, got %s", customPath, service.BasePath)
		}
	})
}

func TestCallOptions(t *testing.T) {
	t.Run("fields option", func(t *testing.T) {
		opts := getCallOptions([]CallOption{Fields("name,size")})
		if opts.fields != "name,size" {
			t.Errorf("Expected fields 'name,size', got '%s'", opts.fields)
		}
	})

	t.Run("no options", func(t *testing.T) {
		opts := getCallOptions(nil)
		if opts.fields != "" {
			t.Errorf("Expected empty fields, got '%s'", opts.fields)
		}
	})
}

func TestMakeRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": {"value": 42}}`))
		}))
		defer server.Close()

		ctx := context.Background()
		service, _ := NewService(ctx,
			WithHTTPClient(server.Client()),
			withBasePath(server.URL+"/"),
		)

		var result map[string]interface{}
		err := service.makeRequest(ctx, "POST", "test", nil, &result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result["result"] == nil {
			t.Error("Expected result in response")
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"code": 400, "message": "Bad request", "status": "INVALID_ARGUMENT"}}`))
		}))
		defer server.Close()

		ctx := context.Background()
		service, _ := NewService(ctx,
			WithHTTPClient(server.Client()),
			withBasePath(server.URL+"/"),
		)

		var result map[string]interface{}
		err := service.makeRequest(ctx, "POST", "test", nil, &result)
		if err == nil {
			t.Fatal("Expected error for API error response")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T", err)
		}
		if apiErr.ErrorInfo.Code != 400 {
			t.Errorf("Expected code 400, got %d", apiErr.ErrorInfo.Code)
		}
	})

	t.Run("with fields option", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query().Get("fields")
			if fields != "name,size" {
				t.Errorf("Expected fields=name,size, got %s", fields)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		ctx := context.Background()
		service, _ := NewService(ctx,
			WithHTTPClient(server.Client()),
			withBasePath(server.URL+"/"),
		)

		var result map[string]interface{}
		err := service.makeRequest(ctx, "GET", "test", nil, &result, Fields("name,size"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}
