package earthengine

import (
	"context"
	"fmt"
)

const (
	// NLCD Tree Canopy Cover dataset (USA only)
	// Using NLCD 2023 TCC - the latest release with annual data from 1985-2023
	// See: https://developers.google.com/earth-engine/datasets/catalog/USGS_NLCD_RELEASES_2023_REL_TCC_v2023-5
	nlcdDatasetID = "USGS/NLCD_RELEASES/2023_REL/TCC/v2023-5"
	nlcdTreeBand  = "NLCD_Percent_Tree_Canopy_Cover"
)

// GetTreeCoverage returns the NLCD tree canopy coverage percentage at the specified point.
// This is a convenience method for quickly getting tree coverage data.
//
// Note: NLCD data only covers the United States. For locations outside the USA,
// this function will return an error or zero coverage.
//
// Parameters:
//   - ctx: Context for the request
//   - latitude: Latitude in decimal degrees (-90 to 90)
//   - longitude: Longitude in decimal degrees (-180 to 180)
//
// Returns:
//   - coverage: Tree canopy coverage as a percentage (0-100)
//   - error: Any error that occurred during the request
func (c *Client) GetTreeCoverage(ctx context.Context, latitude, longitude float64) (float64, error) {
	// Validate coordinates
	if latitude < -90 || latitude > 90 {
		return 0, fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", latitude)
	}
	if longitude < -180 || longitude > 180 {
		return 0, fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", longitude)
	}

	// Create the query using the fluent API
	// Use mosaic to combine all years (later years render on top)
	result, err := c.ImageCollection(nlcdDatasetID).
		Mosaic().
		Select(nlcdTreeBand).
		ReduceRegion(
			NewPoint(longitude, latitude),
			ReducerFirst(),
			Scale(30), // NLCD has 30m resolution
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get tree coverage: %w", err)
	}

	return result, nil
}

// TreeCoverageResult represents the result of a tree coverage query.
type TreeCoverageResult struct {
	Latitude   float64
	Longitude  float64
	Coverage   float64
	DataSource string
}

// GetTreeCoverageDetailed returns detailed tree coverage information including metadata.
func (c *Client) GetTreeCoverageDetailed(ctx context.Context, latitude, longitude float64) (*TreeCoverageResult, error) {
	coverage, err := c.GetTreeCoverage(ctx, latitude, longitude)
	if err != nil {
		return nil, err
	}

	return &TreeCoverageResult{
		Latitude:   latitude,
		Longitude:  longitude,
		Coverage:   coverage,
		DataSource: nlcdDatasetID,
	}, nil
}
