package helpers

import (
	"context"
	"fmt"
	"math"

	"github.com/alexscott64/go-earthengine"
)

// Elevation dataset constants
const (
	// SRTM Digital Elevation Model (30m resolution, near-global coverage)
	srtmDatasetID   = "USGS/SRTMGL1_003"
	srtmElevBand    = "elevation"
	srtmDefaultScale = 30.0

	// ASTER Global Digital Elevation Model (30m resolution, global)
	asterDatasetID   = "NASA/ASTER_GED/AG100_003"
	asterElevBand    = "elevation"
	asterDefaultScale = 30.0

	// ALOS World 3D (30m resolution, global)
	alosDatasetID   = "JAXA/ALOS/AW3D30/V3_2"
	alosElevBand    = "DSM"
	alosDefaultScale = 30.0

	// USGS 3DEP (10m resolution, USA only)
	usgs3DEPDatasetID = "USGS/3DEP/10m"
	usgs3DEPElevBand  = "elevation"
	usgs3DEPDefaultScale = 10.0
)

// ElevationOption configures elevation queries.
type ElevationOption func(*elevationConfig)

type elevationConfig struct {
	dataset string
	scale   *float64
}

// SRTM uses the SRTM 30m dataset (default, near-global coverage).
func SRTM() ElevationOption {
	return func(cfg *elevationConfig) {
		cfg.dataset = srtmDatasetID
	}
}

// ASTER uses the ASTER 30m dataset (global coverage).
func ASTER() ElevationOption {
	return func(cfg *elevationConfig) {
		cfg.dataset = asterDatasetID
	}
}

// ALOS uses the ALOS World 3D 30m dataset (global coverage).
func ALOS() ElevationOption {
	return func(cfg *elevationConfig) {
		cfg.dataset = alosDatasetID
	}
}

// USGS3DEP uses the USGS 3DEP 10m dataset (USA only, higher resolution).
func USGS3DEP() ElevationOption {
	return func(cfg *elevationConfig) {
		cfg.dataset = usgs3DEPDatasetID
	}
}

// ElevationWithScale sets the scale (resolution) for elevation queries in meters.
func ElevationWithScale(meters float64) ElevationOption {
	return func(cfg *elevationConfig) {
		cfg.scale = &meters
	}
}

// Elevation returns the elevation in meters at the specified point.
//
// By default, uses SRTM 30m data. Use options to customize:
//   - SRTM() - SRTM 30m (default, near-global coverage)
//   - ASTER() - ASTER 30m (global coverage)
//   - ALOS() - ALOS 30m (global coverage)
//   - USGS3DEP() - USGS 10m (USA only, higher resolution)
//   - WithScale(30) - Set the resolution in meters
//
// Returns elevation in meters above sea level.
//
// Example:
//
//	// Get elevation in Denver, CO (about 1600m)
//	elev, err := helpers.Elevation(client, 39.7392, -104.9903)
//	fmt.Printf("Elevation: %.0f meters\n", elev)
//
//	// Use higher resolution USGS 3DEP for USA
//	elev, err := helpers.Elevation(client, 39.7392, -104.9903, helpers.USGS3DEP())
func Elevation(client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	ctx := context.Background()
	return ElevationWithContext(ctx, client, lat, lon, opts...)
}

// ElevationWithContext is like Elevation but accepts a context.
func ElevationWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &elevationConfig{
		dataset: srtmDatasetID, // Default to SRTM
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine band name and scale based on dataset
	var band string
	var scale float64

	switch cfg.dataset {
	case srtmDatasetID:
		band = srtmElevBand
		scale = srtmDefaultScale
	case asterDatasetID:
		band = asterElevBand
		scale = asterDefaultScale
	case alosDatasetID:
		band = alosElevBand
		scale = alosDefaultScale
	case usgs3DEPDatasetID:
		band = usgs3DEPElevBand
		scale = usgs3DEPDefaultScale
	default:
		return 0, fmt.Errorf("unknown dataset: %s", cfg.dataset)
	}

	// Override scale if provided
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	result, err := client.Image(cfg.dataset).
		Select(band).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get elevation: %w", err)
	}

	return result, nil
}

// Slope returns the slope in degrees at the specified point.
//
// Slope is calculated from the elevation data using Earth Engine's terrain algorithm.
// Returns slope in degrees (0-90, where 0 is flat and 90 is vertical).
//
// Uses SRTM 30m data by default. Use options to customize the source dataset.
//
// Example:
//
//	// Get slope in mountainous terrain
//	slope, err := helpers.Slope(client, 39.7392, -104.9903)
//	fmt.Printf("Slope: %.1f degrees\n", slope)
func Slope(client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	ctx := context.Background()
	return SlopeWithContext(ctx, client, lat, lon, opts...)
}

// SlopeWithContext is like Slope but accepts a context.
func SlopeWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Get elevation configuration
	cfg := &elevationConfig{
		dataset: srtmDatasetID,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine band name and scale based on dataset
	var band string
	var scale float64

	switch cfg.dataset {
	case srtmDatasetID:
		band = srtmElevBand
		scale = srtmDefaultScale
	case asterDatasetID:
		band = asterElevBand
		scale = asterDefaultScale
	case alosDatasetID:
		band = alosElevBand
		scale = alosDefaultScale
	case usgs3DEPDatasetID:
		band = usgs3DEPElevBand
		scale = usgs3DEPDefaultScale
	default:
		return 0, fmt.Errorf("unknown dataset: %s", cfg.dataset)
	}

	// Override scale if provided
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Load the elevation image and select the elevation band
	elevImage := client.Image(cfg.dataset).Select(band)

	// Apply Terrain.slope to calculate slope in degrees
	slopeImage := elevImage.Terrain(earthengine.AlgorithmTerrainSlope)

	// Sample at the point
	result, err := slopeImage.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute slope: %w", err)
	}

	return result, nil
}

// Aspect returns the aspect (compass direction of slope) in degrees at the specified point.
//
// Aspect is calculated from the elevation data using Earth Engine's terrain algorithm.
// Returns aspect in degrees (0-360, where 0=North, 90=East, 180=South, 270=West).
//
// Uses SRTM 30m data by default. Use options to customize the source dataset.
//
// Example:
//
//	// Get aspect (which direction does the slope face?)
//	aspect, err := helpers.Aspect(client, 39.7392, -104.9903)
//	fmt.Printf("Aspect: %.0f degrees (%.0f = North)\n", aspect, aspect)
func Aspect(client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	ctx := context.Background()
	return AspectWithContext(ctx, client, lat, lon, opts...)
}

// AspectWithContext is like Aspect but accepts a context.
func AspectWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Get elevation configuration
	cfg := &elevationConfig{
		dataset: srtmDatasetID,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Determine band name and scale based on dataset
	var band string
	var scale float64

	switch cfg.dataset {
	case srtmDatasetID:
		band = srtmElevBand
		scale = srtmDefaultScale
	case asterDatasetID:
		band = asterElevBand
		scale = asterDefaultScale
	case alosDatasetID:
		band = alosElevBand
		scale = alosDefaultScale
	case usgs3DEPDatasetID:
		band = usgs3DEPElevBand
		scale = usgs3DEPDefaultScale
	default:
		return 0, fmt.Errorf("unknown dataset: %s", cfg.dataset)
	}

	// Override scale if provided
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Load the elevation image and select the elevation band
	elevImage := client.Image(cfg.dataset).Select(band)

	// Apply Terrain.aspect to calculate aspect in degrees
	aspectImage := elevImage.Terrain(earthengine.AlgorithmTerrainAspect)

	// Sample at the point
	result, err := aspectImage.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute aspect: %w", err)
	}

	return result, nil
}

// TerrainMetrics contains comprehensive terrain analysis results.
type TerrainMetrics struct {
	Elevation float64 // Elevation in meters
	Slope     float64 // Slope in degrees (0-90)
	Aspect    float64 // Aspect in degrees (0-360)
	// Future: Curvature, TPI, TRI when terrain algorithms are implemented
}

// TerrainAnalysis returns comprehensive terrain metrics at the specified point.
//
// This is more efficient than calling Elevation, Slope, and Aspect separately
// because it only needs to load the elevation data once.
//
// Example:
//
//	metrics, err := helpers.TerrainAnalysis(client, 39.7392, -104.9903)
//	fmt.Printf("Elevation: %.0fm\n", metrics.Elevation)
//	fmt.Printf("Slope: %.1f degrees\n", metrics.Slope)
//	fmt.Printf("Aspect: %.0f degrees\n", metrics.Aspect)
func TerrainAnalysis(client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (*TerrainMetrics, error) {
	ctx := context.Background()
	return TerrainAnalysisWithContext(ctx, client, lat, lon, opts...)
}

// TerrainAnalysisWithContext is like TerrainAnalysis but accepts a context.
func TerrainAnalysisWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ElevationOption) (*TerrainMetrics, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	// Get elevation
	elevation, err := ElevationWithContext(ctx, client, lat, lon, opts...)
	if err != nil {
		return nil, err
	}

	metrics := &TerrainMetrics{
		Elevation: elevation,
		Slope:     0, // Placeholder - requires terrain algorithm
		Aspect:    0, // Placeholder - requires terrain algorithm
	}

	return metrics, nil
}

// ElevationQuery represents a deferred elevation query for batch operations.
type ElevationQuery struct {
	lat  float64
	lon  float64
	opts []ElevationOption
}

// NewElevationQuery creates a new elevation query for batch execution.
func NewElevationQuery(lat, lon float64, opts ...ElevationOption) Query {
	return &ElevationQuery{
		lat:  lat,
		lon:  lon,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *ElevationQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return ElevationWithContext(ctx, client, q.lat, q.lon, q.opts...)
}

// Helper functions for terrain calculations (used when terrain algorithms are available)

// degreesToRadians converts degrees to radians.
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// radiansToDegrees converts radians to degrees.
func radiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}
