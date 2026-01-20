package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOperationsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "projects/test/operations/abc123",
			"done": true,
			"response": {"status": "success"}
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	op, err := service.Projects.Operations.Get(ctx, "projects/test/operations/abc123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !op.Done {
		t.Error("Expected operation to be done")
	}
	if op.Name != "projects/test/operations/abc123" {
		t.Errorf("Expected name 'projects/test/operations/abc123', got '%s'", op.Name)
	}
}

func TestOperationsWait(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":wait") {
			t.Errorf("Expected path to contain ':wait', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "projects/test/operations/abc123",
			"done": true,
			"response": {"result": "completed"}
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	req := &WaitOperationRequest{Timeout: "60s"}
	op, err := service.Projects.Operations.Wait(ctx, "projects/test/operations/abc123", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !op.Done {
		t.Error("Expected operation to be done after wait")
	}
}

func TestOperationsWaitWithPolling(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		done := callCount >= 3 // Complete after 3 polls

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if done {
			w.Write([]byte(`{"name": "projects/test/operations/abc123", "done": true}`))
		} else {
			w.Write([]byte(`{"name": "projects/test/operations/abc123", "done": false}`))
		}
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	op, err := service.Projects.Operations.WaitWithPolling(ctx, "projects/test/operations/abc123", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !op.Done {
		t.Error("Expected operation to be done after polling")
	}
	if callCount < 3 {
		t.Errorf("Expected at least 3 calls, got %d", callCount)
	}
}

func TestOperationsCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":cancel") {
			t.Errorf("Expected path to contain ':cancel', got %s", r.URL.Path)
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

	err := service.Projects.Operations.Cancel(ctx, "projects/test/operations/abc123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestOperationsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	err := service.Projects.Operations.Delete(ctx, "projects/test/operations/abc123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
