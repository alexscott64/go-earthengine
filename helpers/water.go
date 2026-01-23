package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// Water dataset constants
const (
	// JRC Global Surface Water - Water occurrence and change (1984-2021, 30m)
	jrcWaterDatasetID = "JRC/GSW1_4/GlobalSurfaceWater"

	// JRC Monthly Water History - Monthly water classification (1984-2021, 30m)
	jrcMonthlyWaterID = "JRC/GSW1_4/MonthlyHistory"
)

// WaterOption configures water queries.
type WaterOption func(*waterConfig)

type waterConfig struct {
	scale float64
}

// WaterWithScale sets the scale for water queries in meters.
func WaterWithScale(meters float64) WaterOption {
	return func(cfg *waterConfig) {
		cfg.scale = meters
	}
}

// WaterDetection checks if a location has water presence.
//
// Returns true if the location has significant water occurrence (>50%).
// Uses JRC Global Surface Water dataset (1984-2021, 30m resolution).
//
// Example:
//
//	hasWater, err := helpers.WaterDetection(client, 45.5152, -122.6784)
//	if hasWater {
//	    fmt.Println("Water detected at location")
//	}
func WaterDetection(client *earthengine.Client, lat, lon float64, opts ...WaterOption) (bool, error) {
	return WaterDetectionWithContext(context.Background(), client, lat, lon, opts...)
}

// WaterDetectionWithContext is like WaterDetection but accepts a context.
func WaterDetectionWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...WaterOption) (bool, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return false, err
	}

	// Apply options
	cfg := &waterConfig{
		scale: 30, // Default to 30m (JRC GSW resolution)
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Query water occurrence
	occurrence, err := WaterOccurrenceWithContext(ctx, client, lat, lon, opts...)
	if err != nil {
		return false, err
	}

	// Consider water present if occurrence > 50%
	return occurrence > 50.0, nil
}

// WaterOccurrence returns the percentage of time water was present at a location.
//
// Returns a value from 0-100 representing the percentage of valid observations
// where water was detected (1984-2021).
//
// Example:
//
//	occurrence, err := helpers.WaterOccurrence(client, 45.5152, -122.6784)
//	fmt.Printf("Water occurrence: %.1f%%\n", occurrence)
func WaterOccurrence(client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	return WaterOccurrenceWithContext(context.Background(), client, lat, lon, opts...)
}

// WaterOccurrenceWithContext is like WaterOccurrence but accepts a context.
func WaterOccurrenceWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &waterConfig{
		scale: 30,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Query JRC Global Surface Water occurrence band
	result, err := client.Image(jrcWaterDatasetID).
		Select("occurrence").
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(cfg.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute water occurrence: %w", err)
	}

	return result, nil
}

// WaterFrequency is an alias for WaterOccurrence for backward compatibility.
func WaterFrequency(client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	return WaterOccurrence(client, lat, lon, opts...)
}

// WaterFrequencyWithContext is an alias for WaterOccurrenceWithContext.
func WaterFrequencyWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	return WaterOccurrenceWithContext(ctx, client, lat, lon, opts...)
}

// WaterSeasonality returns the number of months water is present in a typical year.
//
// Returns a value from 0-12 representing how many months per year water is typically present.
// Higher values indicate more permanent water bodies.
//
// Example:
//
//	seasonality, err := helpers.WaterSeasonality(client, 45.5152, -122.6784)
//	fmt.Printf("Water present %d months/year\n", int(seasonality))
func WaterSeasonality(client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	return WaterSeasonalityWithContext(context.Background(), client, lat, lon, opts...)
}

// WaterSeasonalityWithContext is like WaterSeasonality but accepts a context.
func WaterSeasonalityWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &waterConfig{
		scale: 30,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Query JRC Global Surface Water seasonality band
	result, err := client.Image(jrcWaterDatasetID).
		Select("seasonality").
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(cfg.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute water seasonality: %w", err)
	}

	return result, nil
}

// WaterChange returns the type of water change at a location between epochs.
//
// Returns a value indicating the type of change:
//   - 0: No change
//   - 1: Permanent
//   - 2: New permanent
//   - 3: Lost permanent
//   - 4: Seasonal
//   - 5: New seasonal
//   - 6: Lost seasonal
//   - 7: Seasonal to permanent
//   - 8: Permanent to seasonal
//   - 9: Ephemeral permanent
//   - 10: Ephemeral seasonal
//
// Example:
//
//	change, err := helpers.WaterChange(client, 45.5152, -122.6784)
func WaterChange(client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	return WaterChangeWithContext(context.Background(), client, lat, lon, opts...)
}

// WaterChangeWithContext is like WaterChange but accepts a context.
func WaterChangeWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...WaterOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &waterConfig{
		scale: 30,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Query JRC Global Surface Water change band
	result, err := client.Image(jrcWaterDatasetID).
		Select("change_abs").
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(cfg.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute water change: %w", err)
	}

	return result, nil
}

// WaterDetectionQuery represents a deferred water detection query for batch operations.
type WaterDetectionQuery struct {
	lat  float64
	lon  float64
	opts []WaterOption
}

// NewWaterDetectionQuery creates a new water detection query for batch execution.
func NewWaterDetectionQuery(lat, lon float64, opts ...WaterOption) Query {
	return &WaterDetectionQuery{
		lat:  lat,
		lon:  lon,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *WaterDetectionQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return WaterDetectionWithContext(ctx, client, q.lat, q.lon, q.opts...)
}

// WaterOccurrenceQuery represents a deferred water occurrence query for batch operations.
type WaterOccurrenceQuery struct {
	lat  float64
	lon  float64
	opts []WaterOption
}

// NewWaterOccurrenceQuery creates a new water occurrence query for batch execution.
func NewWaterOccurrenceQuery(lat, lon float64, opts ...WaterOption) Query {
	return &WaterOccurrenceQuery{
		lat:  lat,
		lon:  lon,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *WaterOccurrenceQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return WaterOccurrenceWithContext(ctx, client, q.lat, q.lon, q.opts...)
}
