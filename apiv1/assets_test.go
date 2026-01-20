package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAssetsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "projects/test/assets/my-image",
			"type": "IMAGE",
			"sizeBytes": "1024"
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	asset, err := service.Projects.Assets.Get(ctx, "projects/test/assets/my-image")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if asset.Type != "IMAGE" {
		t.Errorf("Expected type IMAGE, got %s", asset.Type)
	}
	if asset.SizeBytes != 1024 {
		t.Errorf("Expected size 1024, got %d", asset.SizeBytes)
	}
}

func TestAssetsListAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":listAssets") {
			t.Errorf("Expected path to contain ':listAssets', got %s", r.URL.Path)
		}

		pageSize := r.URL.Query().Get("pageSize")
		if pageSize != "10" {
			t.Errorf("Expected pageSize=10, got %s", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"assets": [
				{"name": "projects/test/assets/image1", "type": "IMAGE"},
				{"name": "projects/test/assets/image2", "type": "IMAGE"}
			],
			"nextPageToken": "token123"
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	resp, err := service.Projects.Assets.ListAssets(ctx, "projects/test/assets/folder", 10, "", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(resp.Assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(resp.Assets))
	}
	if resp.NextPageToken != "token123" {
		t.Errorf("Expected nextPageToken 'token123', got '%s'", resp.NextPageToken)
	}
}

func TestAssetsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		assetId := r.URL.Query().Get("assetId")
		if assetId != "my-new-folder" {
			t.Errorf("Expected assetId=my-new-folder, got %s", assetId)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "projects/test/assets/my-new-folder",
			"type": "FOLDER"
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	asset := &EarthEngineAsset{
		Type: "FOLDER",
	}

	created, err := service.Projects.Assets.Create(ctx, "projects/test/assets", "my-new-folder", asset)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if created.Type != "FOLDER" {
		t.Errorf("Expected type FOLDER, got %s", created.Type)
	}
}

func TestAssetsPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}

		updateMask := r.URL.Query().Get("updateMask")
		if updateMask != "title,description" {
			t.Errorf("Expected updateMask=title,description, got %s", updateMask)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "projects/test/assets/my-image",
			"title": "Updated Title"
		}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	asset := &EarthEngineAsset{
		Title:       "Updated Title",
		Description: "Updated description",
	}

	updated, err := service.Projects.Assets.Patch(ctx, "projects/test/assets/my-image", asset, "title,description")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}
}

func TestAssetsDelete(t *testing.T) {
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

	err := service.Projects.Assets.Delete(ctx, "projects/test/assets/my-image")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAssetsCopy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":copy") {
			t.Errorf("Expected path to contain ':copy', got %s", r.URL.Path)
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

	err := service.Projects.Assets.Copy(ctx,
		"projects/test/assets/source",
		"projects/test/assets/dest")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAssetsMove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, ":move") {
			t.Errorf("Expected path to contain ':move', got %s", r.URL.Path)
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

	err := service.Projects.Assets.Move(ctx,
		"projects/test/assets/old-location",
		"projects/test/assets/new-location")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
