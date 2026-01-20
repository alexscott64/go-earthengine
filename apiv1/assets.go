package apiv1

import (
	"context"
	"fmt"
	"net/url"
)

// ProjectsAssetsService handles Earth Engine asset management.
//
// Assets are Earth Engine resources like images, image collections, tables, and folders.
// This service provides full CRUD operations plus IAM policy management.
type ProjectsAssetsService struct {
	s *Service
}

// Get retrieves an asset's metadata.
//
// Example:
//
//	asset, err := service.Projects.Assets.Get(ctx, "projects/my-project/assets/my-folder/my-image")
//	fmt.Printf("Asset type: %s, Size: %d bytes\n", asset.Type, asset.SizeBytes)
func (r *ProjectsAssetsService) Get(ctx context.Context, name string, opts ...CallOption) (*EarthEngineAsset, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	resp := &EarthEngineAsset{}
	if err := r.s.makeRequest(ctx, "GET", name, nil, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// ListAssets lists assets in a folder or collection.
//
// Use pageToken for pagination.
//
// Example:
//
//	resp, err := service.Projects.Assets.ListAssets(ctx, "projects/my-project/assets/my-folder", 100, "")
//	for _, asset := range resp.Assets {
//	    fmt.Printf("Asset: %s (%s)\n", asset.Name, asset.Type)
//	}
func (r *ProjectsAssetsService) ListAssets(ctx context.Context, parent string, pageSize int, pageToken string, filter string, opts ...CallOption) (*ListAssetsResponse, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}

	// Build URL with query parameters
	u, err := url.Parse(r.s.BasePath + parent + ":listAssets")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if pageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	if filter != "" {
		q.Set("filter", filter)
	}
	u.RawQuery = q.Encode()

	urlPath := parent + ":listAssets"
	if q.Encode() != "" {
		urlPath += "?" + q.Encode()
	}

	resp := &ListAssetsResponse{}
	if err := r.s.makeRequest(ctx, "GET", urlPath, nil, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Create creates a new asset.
//
// Example:
//
//	asset := &apiv1.EarthEngineAsset{
//	    Type: "FOLDER",
//	    Name: "projects/my-project/assets/my-new-folder",
//	}
//	created, err := service.Projects.Assets.Create(ctx, "projects/my-project/assets/my-new-folder", asset)
func (r *ProjectsAssetsService) Create(ctx context.Context, parent string, assetId string, asset *EarthEngineAsset, opts ...CallOption) (*EarthEngineAsset, error) {
	if parent == "" {
		return nil, fmt.Errorf("parent is required")
	}
	if assetId == "" {
		return nil, fmt.Errorf("assetId is required")
	}

	urlPath := parent
	u, err := url.Parse(r.s.BasePath + urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("assetId", assetId)
	u.RawQuery = q.Encode()

	urlPath += "?" + q.Encode()

	resp := &EarthEngineAsset{}
	if err := r.s.makeRequest(ctx, "POST", urlPath, asset, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Patch updates an asset's metadata.
//
// Use updateMask to specify which fields to update.
//
// Example:
//
//	asset := &apiv1.EarthEngineAsset{
//	    Title: "Updated Title",
//	    Description: "Updated description",
//	}
//	updated, err := service.Projects.Assets.Patch(ctx, "projects/my-project/assets/my-image", asset, "title,description")
func (r *ProjectsAssetsService) Patch(ctx context.Context, name string, asset *EarthEngineAsset, updateMask string, opts ...CallOption) (*EarthEngineAsset, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	urlPath := name
	if updateMask != "" {
		u, err := url.Parse(r.s.BasePath + urlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		q := u.Query()
		q.Set("updateMask", updateMask)
		u.RawQuery = q.Encode()
		urlPath += "?" + q.Encode()
	}

	resp := &EarthEngineAsset{}
	if err := r.s.makeRequest(ctx, "PATCH", urlPath, asset, resp, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

// Delete deletes an asset.
//
// Example:
//
//	err := service.Projects.Assets.Delete(ctx, "projects/my-project/assets/my-old-image")
func (r *ProjectsAssetsService) Delete(ctx context.Context, name string, opts ...CallOption) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	return r.s.makeRequest(ctx, "DELETE", name, nil, nil, opts...)
}

// Copy copies an asset to a new location.
//
// Example:
//
//	err := service.Projects.Assets.Copy(ctx,
//	    "projects/my-project/assets/source-image",
//	    "projects/my-project/assets/dest-image")
func (r *ProjectsAssetsService) Copy(ctx context.Context, sourceName, destinationName string, opts ...CallOption) error {
	if sourceName == "" {
		return fmt.Errorf("sourceName is required")
	}
	if destinationName == "" {
		return fmt.Errorf("destinationName is required")
	}

	urlPath := sourceName + ":copy"
	req := map[string]string{
		"destinationName": destinationName,
	}

	return r.s.makeRequest(ctx, "POST", urlPath, req, nil, opts...)
}

// Move moves an asset to a new location.
//
// Example:
//
//	err := service.Projects.Assets.Move(ctx,
//	    "projects/my-project/assets/old-location/image",
//	    "projects/my-project/assets/new-location/image")
func (r *ProjectsAssetsService) Move(ctx context.Context, sourceName, destinationName string, opts ...CallOption) error {
	if sourceName == "" {
		return fmt.Errorf("sourceName is required")
	}
	if destinationName == "" {
		return fmt.Errorf("destinationName is required")
	}

	urlPath := sourceName + ":move"
	req := map[string]string{
		"destinationName": destinationName,
	}

	return r.s.makeRequest(ctx, "POST", urlPath, req, nil, opts...)
}
