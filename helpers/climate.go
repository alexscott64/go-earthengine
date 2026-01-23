package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// ClimateOption is a functional option for climate queries.
type ClimateOption func(*ClimateOptions)

// ClimateOptions configures climate data queries.
type ClimateOptions struct {
	dataset   string
	startDate string
	endDate   string
	scale     float64
}

// Climate dataset constants
const (
	// TerraClimate - Monthly climate and water balance (1958-present, 4km)
	terraClimateDatasetID = "IDAHO_EPSCOR/TERRACLIMATE"

	// CHIRPS - Daily precipitation (1981-present, 5km)
	chirpsDatasetID = "UCSB-CHG/CHIRPS/DAILY"

	// SMAP - Soil moisture (2015-present, 9km)
	smapDatasetID = "NASA_USDA/HSL/SMAP_soil_moisture"
)

// TerraClimate uses the TerraClimate monthly dataset (1958-present, 4km).
func TerraClimate() ClimateOption {
	return func(opts *ClimateOptions) {
		opts.dataset = terraClimateDatasetID
		opts.scale = 4000
	}
}

// CHIRPS uses the CHIRPS daily precipitation dataset (1981-present, 5km).
func CHIRPS() ClimateOption {
	return func(opts *ClimateOptions) {
		opts.dataset = chirpsDatasetID
		opts.scale = 5000
	}
}

// SMAP uses the SMAP soil moisture dataset (2015-present, 9km).
func SMAP() ClimateOption {
	return func(opts *ClimateOptions) {
		opts.dataset = smapDatasetID
		opts.scale = 9000
	}
}

// ClimateDateRange sets the date range for climate queries.
func ClimateDateRange(start, end string) ClimateOption {
	return func(opts *ClimateOptions) {
		opts.startDate = start
		opts.endDate = end
	}
}

// Temperature returns the mean temperature at a location for a date range.
//
// Uses TerraClimate by default (monthly, 4km resolution).
// Temperature is returned in degrees Celsius.
//
// Example:
//
//	// Get mean temperature for 2023
//	temp, err := helpers.Temperature(client, 45.5152, -122.6784,
//	    helpers.ClimateDateRange("2023-01-01", "2023-12-31"))
//	fmt.Printf("Mean temperature: %.1f°C\n", temp)
func Temperature(client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	return TemperatureWithContext(context.Background(), client, lat, lon, opts...)
}

// TemperatureWithContext returns the mean temperature with context support.
func TemperatureWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	options := &ClimateOptions{
		dataset: terraClimateDatasetID,
		scale:   4000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Validate date range
	if options.startDate == "" || options.endDate == "" {
		return 0, fmt.Errorf("date range is required (use ClimateDateRange option)")
	}

	// Query using ImageCollection filtering
	result, err := client.ImageCollection(options.dataset).
		FilterDate(options.startDate, options.endDate).
		Select("tmmx"). // Maximum temperature band in TerraClimate
		Reduce(earthengine.ReducerMean()).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(options.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute temperature: %w", err)
	}

	// TerraClimate stores temperature in 0.1°C units, convert to °C
	return result / 10.0, nil
}

// Precipitation returns the total precipitation at a location for a date range.
//
// Uses CHIRPS by default (daily, 5km resolution).
// Precipitation is returned in millimeters.
//
// Example:
//
//	// Get total precipitation for June 2023
//	precip, err := helpers.Precipitation(client, 45.5152, -122.6784,
//	    helpers.ClimateDateRange("2023-06-01", "2023-06-30"))
//	fmt.Printf("Total precipitation: %.1fmm\n", precip)
func Precipitation(client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	return PrecipitationWithContext(context.Background(), client, lat, lon, opts...)
}

// PrecipitationWithContext returns the total precipitation with context support.
func PrecipitationWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	options := &ClimateOptions{
		dataset: chirpsDatasetID,
		scale:   5000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Validate date range
	if options.startDate == "" || options.endDate == "" {
		return 0, fmt.Errorf("date range is required (use ClimateDateRange option)")
	}

	// Query using ImageCollection filtering and sum reducer
	result, err := client.ImageCollection(options.dataset).
		FilterDate(options.startDate, options.endDate).
		Select("precipitation").
		Reduce(earthengine.ReducerSum()).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(options.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute precipitation: %w", err)
	}

	return result, nil
}

// SoilMoisture returns the soil moisture at a location for a date range.
//
// Uses SMAP by default (daily, 9km resolution, 2015-present).
// Soil moisture is returned as volumetric water content (0-1).
//
// Example:
//
//	// Get mean soil moisture for July 2023
//	moisture, err := helpers.SoilMoisture(client, 45.5152, -122.6784,
//	    helpers.ClimateDateRange("2023-07-01", "2023-07-31"))
//	fmt.Printf("Mean soil moisture: %.2f\n", moisture)
func SoilMoisture(client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	return SoilMoistureWithContext(context.Background(), client, lat, lon, opts...)
}

// SoilMoistureWithContext returns the soil moisture with context support.
func SoilMoistureWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, opts ...ClimateOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	options := &ClimateOptions{
		dataset: smapDatasetID,
		scale:   9000,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Validate date range
	if options.startDate == "" || options.endDate == "" {
		return 0, fmt.Errorf("date range is required (use ClimateDateRange option)")
	}

	// Query using ImageCollection filtering
	result, err := client.ImageCollection(options.dataset).
		FilterDate(options.startDate, options.endDate).
		Select("ssm"). // Surface soil moisture
		Reduce(earthengine.ReducerMean()).
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(options.scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute soil moisture: %w", err)
	}

	return result, nil
}

// ClimateQuery represents a climate query for batch processing.
type ClimateQuery struct {
	lat     float64
	lon     float64
	opts    []ClimateOption
	queryFn func(context.Context, *earthengine.Client, float64, float64, ...ClimateOption) (float64, error)
}

// NewTemperatureQuery creates a temperature query for batch processing.
func NewTemperatureQuery(lat, lon float64, opts ...ClimateOption) Query {
	return &ClimateQuery{
		lat:     lat,
		lon:     lon,
		opts:    opts,
		queryFn: TemperatureWithContext,
	}
}

// NewPrecipitationQuery creates a precipitation query for batch processing.
func NewPrecipitationQuery(lat, lon float64, opts ...ClimateOption) Query {
	return &ClimateQuery{
		lat:     lat,
		lon:     lon,
		opts:    opts,
		queryFn: PrecipitationWithContext,
	}
}

// NewSoilMoistureQuery creates a soil moisture query for batch processing.
func NewSoilMoistureQuery(lat, lon float64, opts ...ClimateOption) Query {
	return &ClimateQuery{
		lat:     lat,
		lon:     lon,
		opts:    opts,
		queryFn: SoilMoistureWithContext,
	}
}

// Execute implements the Query interface.
func (q *ClimateQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return q.queryFn(ctx, client, q.lat, q.lon, q.opts...)
}
