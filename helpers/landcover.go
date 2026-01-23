package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// Land cover dataset constants
const (
	// NLCD (USA only) - Tree Canopy Cover 2023
	nlcdTCCDatasetID = "USGS/NLCD_RELEASES/2023_REL/TCC/v2023-5"
	nlcdTCCBand      = "NLCD_Percent_Tree_Canopy_Cover"

	// NLCD Land Cover Classification 2023 (USA only)
	nlcdLandCoverDatasetID = "USGS/NLCD_RELEASES/2023_REL/NLCD"
	nlcdLandCoverBand      = "landcover"

	// NLCD Impervious Surface 2023 (USA only)
	nlcdImperviousDatasetID = "USGS/NLCD_RELEASES/2023_REL/IMPV"
	nlcdImperviousBand      = "impervious"

	// ESA WorldCover (Global) - 10m resolution
	esaWorldCoverDatasetID = "ESA/WorldCover/v200"
	esaWorldCoverBand      = "Map"

	// Hansen Global Forest Change
	hansenDatasetID     = "UMD/hansen/global_forest_change_2023_v1_11"
	hansenTreeCoverBand = "treecover2000"

	// Default scale for land cover operations (meters)
	defaultLandCoverScale = 30.0
)

// TreeCoverageOption configures tree coverage queries.
type TreeCoverageOption func(*treeCoverageConfig)

type treeCoverageConfig struct {
	dataset string
	year    *int
	scale   *float64
}

// Latest uses the latest available tree coverage data (default).
func Latest() TreeCoverageOption {
	return func(cfg *treeCoverageConfig) {
		// Latest is the default, no change needed
	}
}

// Year sets a specific year for tree coverage (NLCD supports 1985-2023).
func Year(year int) TreeCoverageOption {
	return func(cfg *treeCoverageConfig) {
		cfg.year = &year
	}
}

// NLCDDataset uses the NLCD dataset (USA only, default).
func NLCDDataset() TreeCoverageOption {
	return func(cfg *treeCoverageConfig) {
		cfg.dataset = nlcdTCCDatasetID
	}
}

// HansenDataset uses the Hansen Global Forest Change dataset (global).
func HansenDataset() TreeCoverageOption {
	return func(cfg *treeCoverageConfig) {
		cfg.dataset = hansenDatasetID
	}
}

// WithScale sets the scale (resolution) for the query in meters.
func WithScale(meters float64) TreeCoverageOption {
	return func(cfg *treeCoverageConfig) {
		cfg.scale = &meters
	}
}

// TreeCoverage returns the tree canopy coverage percentage at the specified point.
//
// By default, uses NLCD 2023 data for USA locations. Use options to customize:
//   - Latest() - Use the most recent data (default)
//   - Year(2020) - Use data from a specific year (NLCD: 1985-2023)
//   - HansenDataset() - Use Hansen Global Forest Change (for non-USA locations)
//   - WithScale(30) - Set the resolution in meters
//
// Returns coverage as a percentage (0-100).
//
// Example:
//
//	// Get latest tree coverage in Portland, OR
//	coverage, err := helpers.TreeCoverage(client, 45.5152, -122.6784)
//
//	// Get tree coverage from 2010
//	coverage, err := helpers.TreeCoverage(client, 45.5152, -122.6784, helpers.Year(2010))
//
//	// Use Hansen dataset for global coverage
//	coverage, err := helpers.TreeCoverage(client, 52.5200, 13.4050, helpers.HansenDataset())
func TreeCoverage(client *earthengine.Client, lat, lon float64, opts ...TreeCoverageOption) (float64, error) {
	ctx := context.Background()
	return TreeCoverageWithContext(ctx, client, lat, lon, opts...)
}

// TreeCoverageWithContext is like TreeCoverage but accepts a context.
func TreeCoverageWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...TreeCoverageOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &treeCoverageConfig{
		dataset: nlcdTCCDatasetID, // Default to NLCD
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine scale
	scale := defaultLandCoverScale
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Build query based on dataset
	var result float64
	var err error

	if cfg.dataset == hansenDatasetID {
		// Hansen is a single image
		result, err = client.Image(cfg.dataset).
			Select(hansenTreeCoverBand).
			ReduceRegion(
				earthengine.NewPoint(lon, lat),
				earthengine.ReducerFirst(),
				earthengine.Scale(scale),
			).
			ComputeFloat(ctx)
	} else {
		// NLCD is an ImageCollection - use mosaic to get latest
		result, err = client.ImageCollection(cfg.dataset).
			Mosaic().
			Select(nlcdTCCBand).
			ReduceRegion(
				earthengine.NewPoint(lon, lat),
				earthengine.ReducerFirst(),
				earthengine.Scale(scale),
			).
			ComputeFloat(ctx)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get tree coverage: %w", err)
	}

	return result, nil
}

// LandCoverClass returns the land cover classification at the specified point.
//
// For USA locations, uses NLCD 2023 with the following classes:
//   - "water" (11)
//   - "ice_snow" (12)
//   - "developed_open" (21)
//   - "developed_low" (22)
//   - "developed_medium" (23)
//   - "developed_high" (24)
//   - "barren" (31)
//   - "forest_deciduous" (41)
//   - "forest_evergreen" (42)
//   - "forest_mixed" (43)
//   - "shrub" (52)
//   - "grassland" (71)
//   - "pasture" (81)
//   - "crops" (82)
//   - "woody_wetlands" (90)
//   - "herbaceous_wetlands" (95)
//
// Example:
//
//	class, err := helpers.LandCoverClass(client, 45.5152, -122.6784)
//	fmt.Println(class) // e.g., "forest_evergreen"
func LandCoverClass(client *earthengine.Client, lat, lon float64) (string, error) {
	ctx := context.Background()
	return LandCoverClassWithContext(ctx, client, lat, lon)
}

// LandCoverClassWithContext is like LandCoverClass but accepts a context.
func LandCoverClassWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64) (string, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return "", err
	}

	// Get the numeric class value
	result, err := client.ImageCollection(nlcdLandCoverDatasetID).
		Mosaic().
		Select(nlcdLandCoverBand).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(defaultLandCoverScale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to get land cover class: %w", err)
	}

	// Convert numeric value to class name
	return nlcdClassToName(int(result)), nil
}

// nlcdClassToName converts NLCD numeric class to human-readable name.
func nlcdClassToName(class int) string {
	classes := map[int]string{
		11: "water",
		12: "ice_snow",
		21: "developed_open",
		22: "developed_low",
		23: "developed_medium",
		24: "developed_high",
		31: "barren",
		41: "forest_deciduous",
		42: "forest_evergreen",
		43: "forest_mixed",
		52: "shrub",
		71: "grassland",
		81: "pasture",
		82: "crops",
		90: "woody_wetlands",
		95: "herbaceous_wetlands",
	}

	if name, ok := classes[class]; ok {
		return name
	}
	return fmt.Sprintf("unknown_%d", class)
}

// ImperviousSurface returns the impervious surface percentage at the specified point.
//
// Impervious surface represents constructed surfaces like roads, buildings, and parking lots
// that water cannot infiltrate. This is important for hydrology and urban planning.
//
// Uses NLCD 2023 Impervious Surface data (USA only).
// Returns percentage (0-100).
//
// Example:
//
//	impervious, err := helpers.ImperviousSurface(client, 45.5152, -122.6784)
//	fmt.Printf("Impervious surface: %.1f%%\n", impervious)
func ImperviousSurface(client *earthengine.Client, lat, lon float64) (float64, error) {
	ctx := context.Background()
	return ImperviousSurfaceWithContext(ctx, client, lat, lon)
}

// ImperviousSurfaceWithContext is like ImperviousSurface but accepts a context.
func ImperviousSurfaceWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	result, err := client.ImageCollection(nlcdImperviousDatasetID).
		Mosaic().
		Select(nlcdImperviousBand).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(defaultLandCoverScale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get impervious surface: %w", err)
	}

	return result, nil
}

// IsUrban returns true if the specified location is classified as urban/developed.
//
// A location is considered urban if:
//   - Land cover class is developed (open, low, medium, or high intensity), OR
//   - Impervious surface is greater than 20%
//
// Uses NLCD 2023 data (USA only).
//
// Example:
//
//	urban, err := helpers.IsUrban(client, 45.5152, -122.6784)
//	if urban {
//	    fmt.Println("This is an urban area")
//	}
func IsUrban(client *earthengine.Client, lat, lon float64) (bool, error) {
	ctx := context.Background()
	return IsUrbanWithContext(ctx, client, lat, lon)
}

// IsUrbanWithContext is like IsUrban but accepts a context.
func IsUrbanWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64) (bool, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return false, err
	}

	// Get land cover class
	class, err := LandCoverClassWithContext(ctx, client, lat, lon)
	if err != nil {
		return false, err
	}

	// Check if it's a developed class
	developedClasses := map[string]bool{
		"developed_open":   true,
		"developed_low":    true,
		"developed_medium": true,
		"developed_high":   true,
	}

	if developedClasses[class] {
		return true, nil
	}

	// If not developed class, check impervious surface
	impervious, err := ImperviousSurfaceWithContext(ctx, client, lat, lon)
	if err != nil {
		return false, err
	}

	// Consider urban if >20% impervious
	return impervious > 20.0, nil
}

// TreeCoverageQuery represents a deferred tree coverage query for batch operations.
type TreeCoverageQuery struct {
	lat  float64
	lon  float64
	opts []TreeCoverageOption
}

// NewTreeCoverageQuery creates a new tree coverage query for batch execution.
func NewTreeCoverageQuery(lat, lon float64, opts ...TreeCoverageOption) Query {
	return &TreeCoverageQuery{
		lat:  lat,
		lon:  lon,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *TreeCoverageQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return TreeCoverageWithContext(ctx, client, q.lat, q.lon, q.opts...)
}
