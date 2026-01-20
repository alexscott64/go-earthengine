package apiv1

import (
	"context"
	"fmt"
)

// ProjectsService provides access to project-scoped API resources.
type ProjectsService struct {
	s *Service

	// Sub-resources
	Value           *ProjectsValueService
	Image           *ProjectsImageService
	Table           *ProjectsTableService
	Operations      *ProjectsOperationsService
	Assets          *ProjectsAssetsService
	ImageCollection *ProjectsImageCollectionService
	Thumbnails      *ProjectsThumbnailsService
	Maps            *ProjectsMapsService
	Algorithms      *ProjectsAlgorithmsService
}

// ===== Projects.Value Service =====

// ProjectsValueService handles single value computation.
type ProjectsValueService struct {
	s *Service
}

// Compute evaluates an Earth Engine expression and returns a single value.
//
// This is the method we've been using via the legacy client.ComputeValue.
// It's the core method for evaluating expressions.
//
// Example:
//
//	req := &apiv1.ComputeValueRequest{
//	    Expression: expr,
//	}
//	resp, err := service.Projects.Value.Compute(ctx, "projects/my-project", req)
func (r *ProjectsValueService) Compute(ctx context.Context, parent string, req *ComputeValueRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/value:compute", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.Image Service =====

// ProjectsImageService handles image operations.
type ProjectsImageService struct {
	s *Service
}

// ComputePixels computes and returns pixel values from an image.
//
// This method returns raw pixel data in the specified format (GeoTIFF, NPY, etc.).
//
// Example:
//
//	req := &apiv1.ComputePixelsRequest{
//	    Expression: imageExpr,
//	    FileFormat: "GEO_TIFF",
//	    Grid: grid,
//	}
//	resp, err := service.Projects.Image.ComputePixels(ctx, "projects/my-project", req)
func (r *ProjectsImageService) ComputePixels(ctx context.Context, parent string, req *ComputePixelsRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/image:computePixels", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Export exports an image to Google Drive or Cloud Storage.
//
// This method returns an Operation that tracks the export progress.
//
// Example:
//
//	req := &apiv1.ExportImageRequest{
//	    Expression: imageExpr,
//	    FileExportOptions: &apiv1.FileExportOptions{
//	        FileFormat: "GEO_TIFF",
//	        GcsDestination: &apiv1.GcsDestination{
//	            Bucket: "my-bucket",
//	            FilenamePrefix: "output",
//	        },
//	    },
//	}
//	op, err := service.Projects.Image.Export(ctx, "projects/my-project", req)
//	// Wait for export to complete
//	op, err = service.Projects.Operations.Wait(ctx, op.Name, &apiv1.WaitOperationRequest{Timeout: "3600s"})
func (r *ProjectsImageService) Export(ctx context.Context, parent string, req *ExportImageRequest, opts ...CallOption) (*Operation, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/image:export", parent)
	resp := &Operation{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.Table Service =====

// ProjectsTableService handles table/vector operations.
type ProjectsTableService struct {
	s *Service
}

// ComputeFeatures computes features from a feature collection expression.
//
// Returns features in GeoJSON, CSV, or other formats.
//
// Example:
//
//	req := &apiv1.ComputeFeaturesRequest{
//	    Expression: tableExpr,
//	    FileFormat: "GEO_JSON",
//	}
//	resp, err := service.Projects.Table.ComputeFeatures(ctx, "projects/my-project", req)
func (r *ProjectsTableService) ComputeFeatures(ctx context.Context, parent string, req *ComputeFeaturesRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/table:computeFeatures", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Export exports a feature collection to Google Drive or Cloud Storage.
func (r *ProjectsTableService) Export(ctx context.Context, parent string, req *ExportImageRequest, opts ...CallOption) (*Operation, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/table:export", parent)
	resp := &Operation{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.ImageCollection Service =====

// ProjectsImageCollectionService handles image collection operations.
type ProjectsImageCollectionService struct {
	s *Service
}

// ComputeImages computes multiple images from an ImageCollection.
func (r *ProjectsImageCollectionService) ComputeImages(ctx context.Context, parent string, req *ComputePixelsRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/imageCollection:computeImages", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.Thumbnails Service =====

// ProjectsThumbnailsService handles thumbnail generation.
type ProjectsThumbnailsService struct {
	s *Service
}

// Create generates a thumbnail of an image.
//
// Returns a URL to the thumbnail image.
func (r *ProjectsThumbnailsService) Create(ctx context.Context, parent string, req *ComputePixelsRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/thumbnails", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.Maps Service =====

// ProjectsMapsService handles map tile generation.
type ProjectsMapsService struct {
	s *Service
}

// Create creates a map visualization.
func (r *ProjectsMapsService) Create(ctx context.Context, parent string, req *ComputePixelsRequest, opts ...CallOption) (*ComputeValueResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/maps", parent)
	resp := &ComputeValueResponse{}

	if err := r.s.makeRequest(ctx, "POST", urlPath, req, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ===== Projects.Algorithms Service =====

// ProjectsAlgorithmsService lists available algorithms.
type ProjectsAlgorithmsService struct {
	s *Service
}

// List returns the list of available Earth Engine algorithms.
//
// This is useful for discovering what functions are available.
//
// Example:
//
//	resp, err := service.Projects.Algorithms.List(ctx, "projects/my-project")
//	for _, algo := range resp.Algorithms {
//	    fmt.Println(algo)
//	}
func (r *ProjectsAlgorithmsService) List(ctx context.Context, parent string, opts ...CallOption) (*ListAlgorithmsResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	urlPath := fmt.Sprintf("%s/algorithms", parent)
	resp := &ListAlgorithmsResponse{}

	if err := r.s.makeRequest(ctx, "GET", urlPath, nil, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}
