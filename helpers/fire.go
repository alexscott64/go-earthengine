package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// Fire dataset constants
const (
	// VIIRS Active Fire - Near real-time fire detection (375m, daily)
	viirsFireDatasetID = "FIRMS"

	// MODIS Active Fire - Fire detection (1km, daily)
	modisFireDatasetID = "MODIS/006/MOD14A1"

	// Landsat 8 for burn severity analysis
	landsat8SRID = "LANDSAT/LC08/C02/T1_L2"
)

// FireOption configures fire queries.
type FireOption func(*fireConfig)

type fireConfig struct {
	dataset   string
	dateRange *DateRange
	scale     float64
}

// VIIRS uses VIIRS 375m active fire dataset.
func VIIRS() FireOption {
	return func(cfg *fireConfig) {
		cfg.dataset = viirsFireDatasetID
		cfg.scale = 375
	}
}

// MODISFire uses MODIS 1km active fire dataset.
func MODISFire() FireOption {
	return func(cfg *fireConfig) {
		cfg.dataset = modisFireDatasetID
		cfg.scale = 1000
	}
}

// FireDateRange sets the date range for fire queries.
func FireDateRange(start, end string) FireOption {
	return func(cfg *fireConfig) {
		cfg.dateRange = &DateRange{Start: start, End: end}
	}
}

// FireWithScale sets the scale for fire queries in meters.
func FireWithScale(meters float64) FireOption {
	return func(cfg *fireConfig) {
		cfg.scale = meters
	}
}

// ActiveFire detects if there are active fires at a location within a date range.
//
// Returns true if any fire detections occurred at the location.
// Uses MODIS active fire dataset by default (1km resolution, daily).
//
// Example:
//
//	hasActiveFire, err := helpers.ActiveFire(client, 45.5152, -122.6784,
//	    helpers.FireDateRange("2023-08-01", "2023-08-31"))
func ActiveFire(client *earthengine.Client, lat, lon float64, opts ...FireOption) (bool, error) {
	return ActiveFireWithContext(context.Background(), client, lat, lon, opts...)
}

// ActiveFireWithContext is like ActiveFire but accepts a context.
func ActiveFireWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...FireOption) (bool, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return false, err
	}

	// Apply options
	cfg := &fireConfig{
		dataset: modisFireDatasetID,
		scale:   1000,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.dateRange == nil {
		return false, fmt.Errorf("date range is required (use FireDateRange option)")
	}

	// Query fire detections
	count, err := FireCountWithContext(ctx, client, lat, lon, opts...)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FireCount counts the number of fire detections at a location within a date range.
//
// Returns the total number of fire detection events.
//
// Example:
//
//	count, err := helpers.FireCount(client, 45.5152, -122.6784,
//	    helpers.FireDateRange("2023-01-01", "2023-12-31"))
//	fmt.Printf("Fire detections: %d\n", int(count))
func FireCount(client *earthengine.Client, lat, lon float64, opts ...FireOption) (float64, error) {
	return FireCountWithContext(context.Background(), client, lat, lon, opts...)
}

// FireCountWithContext is like FireCount but accepts a context.
func FireCountWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...FireOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &fireConfig{
		dataset: modisFireDatasetID,
		scale:   1000,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.dateRange == nil {
		return 0, fmt.Errorf("date range is required (use FireDateRange option)")
	}

	// Query MODIS fire dataset
	result, err := client.ImageCollection(cfg.dataset).
		FilterDate(cfg.dateRange.Start, cfg.dateRange.End).
		Select("MaxFRP"). // Maximum Fire Radiative Power
		Count().
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(cfg.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute fire count: %w", err)
	}

	return result, nil
}

// BurnSeverity calculates the Normalized Burn Ratio (NBR) for a location.
//
// NBR = (NIR - SWIR) / (NIR + SWIR)
//
// NBR is used to identify burned areas and assess burn severity.
// Higher values (> 0.1) indicate healthy vegetation.
// Lower or negative values indicate bare ground or burned areas.
//
// Example:
//
//	nbr, err := helpers.BurnSeverity(client, 45.5152, -122.6784, "2023-08-15",
//	    helpers.Landsat8())
func BurnSeverity(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	return BurnSeverityWithContext(context.Background(), client, lat, lon, date, opts...)
}

// BurnSeverityWithContext is like BurnSeverity but accepts a context.
func BurnSeverityWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get NIR (B5) and SWIR2 (B7) bands for Landsat 8
	nirBand := "SR_B5"
	swirBand := "SR_B7"

	if cfg.dataset == sentinel2DatasetID {
		nirBand = "B8"
		swirBand = "B12" // Sentinel-2 SWIR2
	}

	// Build the query
	collection := client.ImageCollection(cfg.dataset)

	// Apply date filtering
	if cfg.dateRange != nil {
		collection = collection.FilterDate(cfg.dateRange.Start, cfg.dateRange.End)
	} else {
		collection = collection.FilterDate(date, date)
	}

	// Apply cloud filtering if specified
	if cfg.cloudCover != nil {
		collection = collection.FilterMetadata("CLOUD_COVER", "less_than", *cfg.cloudCover)
	}

	// Calculate NBR = (NIR - SWIR) / (NIR + SWIR)
	image := collection.
		Select(nirBand, swirBand).
		Reduce(earthengine.ReducerMean()).
		NormalizedDifference()

	// Determine scale
	scale := defaultImageryScale
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Sample at the point
	result, err := image.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute burn severity: %w", err)
	}

	return result, nil
}

// DeltaNBR calculates the difference in NBR between pre-fire and post-fire images.
//
// dNBR = NBR(pre-fire) - NBR(post-fire)
//
// Burn severity classification:
//   - dNBR < 0.1: Unburned
//   - 0.1 - 0.27: Low severity
//   - 0.27 - 0.44: Moderate-low severity
//   - 0.44 - 0.66: Moderate-high severity
//   - dNBR > 0.66: High severity
//
// Example:
//
//	dnbr, err := helpers.DeltaNBR(client, 45.5152, -122.6784,
//	    "2023-07-01", "2023-09-01",
//	    helpers.Landsat8())
func DeltaNBR(client *earthengine.Client, lat, lon float64, preFireDate, postFireDate string, opts ...ImageryOption) (float64, error) {
	return DeltaNBRWithContext(context.Background(), client, lat, lon, preFireDate, postFireDate, opts...)
}

// DeltaNBRWithContext is like DeltaNBR but accepts a context.
func DeltaNBRWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, preFireDate, postFireDate string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Get pre-fire NBR
	preNBR, err := BurnSeverityWithContext(ctx, client, lat, lon, preFireDate, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to compute pre-fire NBR: %w", err)
	}

	// Get post-fire NBR
	postNBR, err := BurnSeverityWithContext(ctx, client, lat, lon, postFireDate, opts...)
	if err != nil {
		return 0, fmt.Errorf("failed to compute post-fire NBR: %w", err)
	}

	// Calculate dNBR = pre - post
	return preNBR - postNBR, nil
}

// ActiveFireQuery represents a deferred active fire query for batch operations.
type ActiveFireQuery struct {
	lat  float64
	lon  float64
	opts []FireOption
}

// NewActiveFireQuery creates a new active fire query for batch execution.
func NewActiveFireQuery(lat, lon float64, opts ...FireOption) Query {
	return &ActiveFireQuery{
		lat:  lat,
		lon:  lon,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *ActiveFireQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return ActiveFireWithContext(ctx, client, q.lat, q.lon, q.opts...)
}

// BurnSeverityQuery represents a deferred burn severity query for batch operations.
type BurnSeverityQuery struct {
	lat  float64
	lon  float64
	date string
	opts []ImageryOption
}

// NewBurnSeverityQuery creates a new burn severity query for batch execution.
func NewBurnSeverityQuery(lat, lon float64, date string, opts ...ImageryOption) Query {
	return &BurnSeverityQuery{
		lat:  lat,
		lon:  lon,
		date: date,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *BurnSeverityQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return BurnSeverityWithContext(ctx, client, q.lat, q.lon, q.date, q.opts...)
}
