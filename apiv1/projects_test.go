package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectsValueCompute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "value:compute") {
			t.Errorf("Expected path to contain 'value:compute', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": {"tree_canopy": 72}}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	req := &ComputeValueRequest{
		Expression: &Expression{
			Result: "0",
			Values: map[string]*ValueNode{
				"0": {ConstantValue: 42},
			},
		},
	}

	resp, err := service.Projects.Value.Compute(ctx, "projects/test-project", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map, got %T", resp.Result)
	}

	if result["tree_canopy"] != float64(72) {
		t.Errorf("Expected tree_canopy=72, got %v", result["tree_canopy"])
	}
}

func TestProjectsValueCompute_MissingParent(t *testing.T) {
	ctx := context.Background()
	service, _ := NewService(ctx, WithHTTPClient(http.DefaultClient))

	req := &ComputeValueRequest{
		Expression: &Expression{},
	}

	_, err := service.Projects.Value.Compute(ctx, "", req)
	if err == nil {
		t.Error("Expected error for missing parent")
	}
}

func TestProjectsImageExport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "image:export") {
			t.Errorf("Expected path to contain 'image:export', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "projects/test/operations/abc123", "done": false}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	req := &ExportImageRequest{
		Expression: &Expression{
			Result: "0",
			Values: map[string]*ValueNode{},
		},
		FileExportOptions: &FileExportOptions{
			FileFormat: "GEO_TIFF",
			GcsDestination: &GcsDestination{
				Bucket:         "my-bucket",
				FilenamePrefix: "output",
			},
		},
	}

	op, err := service.Projects.Image.Export(ctx, "projects/test-project", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if op.Name != "projects/test/operations/abc123" {
		t.Errorf("Expected operation name 'projects/test/operations/abc123', got '%s'", op.Name)
	}
	if op.Done {
		t.Error("Expected operation to not be done")
	}
}

func TestProjectsTableComputeFeatures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "table:computeFeatures") {
			t.Errorf("Expected path to contain 'table:computeFeatures', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": {"type": "FeatureCollection", "features": []}}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	req := &ComputeFeaturesRequest{
		Expression: &Expression{},
		FileFormat: "GEO_JSON",
	}

	resp, err := service.Projects.Table.ComputeFeatures(ctx, "projects/test-project", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Result == nil {
		t.Error("Expected result in response")
	}
}

func TestProjectsAlgorithmsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "algorithms") {
			t.Errorf("Expected path to contain 'algorithms', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"algorithms": ["Image.load", "Image.select", "Reducer.first"]}`))
	}))
	defer server.Close()

	ctx := context.Background()
	service, _ := NewService(ctx,
		WithHTTPClient(server.Client()),
		withBasePath(server.URL+"/"),
	)

	resp, err := service.Projects.Algorithms.List(ctx, "projects/test-project")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(resp.Algorithms) != 3 {
		t.Errorf("Expected 3 algorithms, got %d", len(resp.Algorithms))
	}

	if resp.Algorithms[0] != "Image.load" {
		t.Errorf("Expected first algorithm to be 'Image.load', got '%s'", resp.Algorithms[0])
	}
}
