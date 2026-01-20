package helpers

import (
	"context"
	"fmt"

	"github.com/yourusername/go-earthengine"
)

// Imagery dataset constants
const (
	// Landsat 8 Collection 2 Level 2 (surface reflectance)
	landsat8DatasetID = "LANDSAT/LC08/C02/T1_L2"

	// Landsat 9 Collection 2 Level 2 (surface reflectance)
	landsat9DatasetID = "LANDSAT/LC09/C02/T1_L2"

	// Sentinel-2 Level 2A (surface reflectance)
	sentinel2DatasetID = "COPERNICUS/S2_SR_HARMONIZED"

	// MODIS Terra Vegetation Indices (NDVI, EVI)
	modisVIDatasetID = "MODIS/006/MOD13A1"

	// Default scale for imagery operations (meters)
	defaultImageryScale = 30.0
)

// ImageryOption configures imagery queries.
type ImageryOption func(*imageryConfig)

type imageryConfig struct {
	dataset    string
	cloudCover *float64
	dateRange  *DateRange
	scale      *float64
}

// Landsat8 uses Landsat 8 imagery (default, 30m resolution).
func Landsat8() ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.dataset = landsat8DatasetID
	}
}

// Landsat9 uses Landsat 9 imagery (30m resolution).
func Landsat9() ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.dataset = landsat9DatasetID
	}
}

// Sentinel2 uses Sentinel-2 imagery (10-20m resolution).
func Sentinel2() ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.dataset = sentinel2DatasetID
	}
}

// MODIS uses MODIS vegetation indices (250m resolution).
func MODIS() ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.dataset = modisVIDatasetID
	}
}

// CloudMask sets the maximum cloud cover percentage (0-100).
func CloudMask(maxCloudPercent float64) ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.cloudCover = &maxCloudPercent
	}
}

// DateRangeOption sets the date range for imagery queries.
func DateRangeOption(start, end string) ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.dateRange = &DateRange{Start: start, End: end}
	}
}

// ImageryWithScale sets the scale (resolution) for imagery queries in meters.
func ImageryWithScale(meters float64) ImageryOption {
	return func(cfg *imageryConfig) {
		cfg.scale = &meters
	}
}

// getBandNames returns the NIR and Red band names for a given dataset.
func getBandNames(dataset string) (nir, red string) {
	switch dataset {
	case landsat8DatasetID, landsat9DatasetID:
		return "SR_B5", "SR_B4" // Landsat 8/9: B5=NIR, B4=Red
	case sentinel2DatasetID:
		return "B8", "B4" // Sentinel-2: B8=NIR, B4=Red
	case modisVIDatasetID:
		return "sur_refl_b02", "sur_refl_b01" // MODIS: b02=NIR, b01=Red
	default:
		return "SR_B5", "SR_B4" // Default to Landsat
	}
}

// getBandNamesForWater returns the Green and NIR band names for NDWI calculation.
func getBandNamesForWater(dataset string) (green, nir string) {
	switch dataset {
	case landsat8DatasetID, landsat9DatasetID:
		return "SR_B3", "SR_B5" // Landsat: B3=Green, B5=NIR
	case sentinel2DatasetID:
		return "B3", "B8" // Sentinel-2: B3=Green, B8=NIR
	default:
		return "SR_B3", "SR_B5" // Default to Landsat
	}
}

// getBandNamesForBuiltUp returns the SWIR and NIR band names for NDBI calculation.
func getBandNamesForBuiltUp(dataset string) (swir, nir string) {
	switch dataset {
	case landsat8DatasetID, landsat9DatasetID:
		return "SR_B6", "SR_B5" // Landsat: B6=SWIR1, B5=NIR
	case sentinel2DatasetID:
		return "B11", "B8" // Sentinel-2: B11=SWIR, B8=NIR
	default:
		return "SR_B6", "SR_B5" // Default to Landsat
	}
}

// getBandNamesForEVI returns the NIR, Red, and Blue band names for EVI calculation.
func getBandNamesForEVI(dataset string) (nir, red, blue string) {
	switch dataset {
	case landsat8DatasetID, landsat9DatasetID:
		return "SR_B5", "SR_B4", "SR_B2" // Landsat: B5=NIR, B4=Red, B2=Blue
	case sentinel2DatasetID:
		return "B8", "B4", "B2" // Sentinel-2: B8=NIR, B4=Red, B2=Blue
	default:
		return "SR_B5", "SR_B4", "SR_B2" // Default to Landsat
	}
}

// NDVI calculates the Normalized Difference Vegetation Index at a point.
//
// NDVI = (NIR - Red) / (NIR + Red)
// Values range from -1 to 1, where:
//   - < 0: Water, clouds, snow
//   - 0-0.2: Bare soil, rock
//   - 0.2-0.5: Sparse vegetation, grassland
//   - 0.5-0.8: Dense vegetation, forest
//   - > 0.8: Very dense vegetation
//
// Example:
//
//	// Get NDVI for a location in summer 2023
//	ndvi, err := helpers.NDVI(client, 45.5152, -122.6784, "2023-06-01",
//	    helpers.Sentinel2(),
//	    helpers.DateRangeOption("2023-06-01", "2023-08-31"),
//	    helpers.CloudMask(20))
func NDVI(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	ctx := context.Background()
	return NDVIWithContext(ctx, client, lat, lon, date, opts...)
}

// NDVIWithContext is like NDVI but accepts a context.
func NDVIWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the appropriate band names for NIR and Red
	nirBand, redBand := getBandNames(cfg.dataset)

	// Build the query
	collection := client.ImageCollection(cfg.dataset)

	// Apply date filtering
	if cfg.dateRange != nil {
		collection = collection.FilterDate(cfg.dateRange.Start, cfg.dateRange.End)
	} else {
		// Use a 30-day window centered on the date
		collection = collection.FilterDate(date, date)
	}

	// Apply cloud filtering if specified
	if cfg.cloudCover != nil {
		collection = collection.FilterMetadata("CLOUD_COVER", "less_than", *cfg.cloudCover)
	}

	// Select NIR and Red bands, calculate NDVI using normalized difference
	image := collection.
		Select(nirBand, redBand).
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
		return 0, fmt.Errorf("failed to compute NDVI: %w", err)
	}

	return result, nil
}

// EVI calculates the Enhanced Vegetation Index at a point.
//
// EVI = 2.5 * ((NIR - Red) / (NIR + 6*Red - 7.5*Blue + 1))
//
// EVI is more sensitive to canopy structure and reduces atmospheric influences
// compared to NDVI. Values range from -1 to 1.
//
// Example:
//
//	evi, err := helpers.EVI(client, 45.5152, -122.6784, "2023-06-01",
//	    helpers.Sentinel2())
func EVI(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	ctx := context.Background()
	return EVIWithContext(ctx, client, lat, lon, date, opts...)
}

// EVIWithContext is like EVI but accepts a context.
func EVIWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the appropriate band names
	nirBand, redBand, blueBand := getBandNamesForEVI(cfg.dataset)

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

	// Get mean image and select required bands
	image := collection.Select(nirBand, redBand, blueBand).Reduce(earthengine.ReducerMean())

	// Calculate EVI using expression: 2.5 * ((NIR - Red) / (NIR + 6*Red - 7.5*Blue + 1))
	evi := image.Expression("2.5 * ((NIR - RED) / (NIR + 6*RED - 7.5*BLUE + 1))",
		map[string]interface{}{
			"NIR":  image.Select(nirBand + "_mean"),
			"RED":  image.Select(redBand + "_mean"),
			"BLUE": image.Select(blueBand + "_mean"),
		})

	// Determine scale
	scale := defaultImageryScale
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Sample at the point
	result, err := evi.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute EVI: %w", err)
	}

	return result, nil
}

// SAVI calculates the Soil-Adjusted Vegetation Index at a point.
//
// SAVI = ((NIR - Red) / (NIR + Red + L)) * (1 + L)
// where L is a soil brightness correction factor (typically 0.5)
//
// SAVI is useful in areas with sparse vegetation where soil background
// affects the reflectance.
//
// Example:
//
//	savi, err := helpers.SAVI(client, 45.5152, -122.6784, "2023-06-01")
func SAVI(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	ctx := context.Background()
	return SAVIWithContext(ctx, client, lat, lon, date, opts...)
}

// SAVIWithContext is like SAVI but accepts a context.
func SAVIWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the appropriate band names
	nirBand, redBand := getBandNames(cfg.dataset)

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

	// Get mean image and select required bands
	image := collection.Select(nirBand, redBand).Reduce(earthengine.ReducerMean())

	// Calculate SAVI using expression: ((NIR - Red) / (NIR + Red + L)) * (1 + L)
	// L = 0.5 (soil brightness correction factor)
	savi := image.Expression("((NIR - RED) / (NIR + RED + 0.5)) * 1.5",
		map[string]interface{}{
			"NIR": image.Select(nirBand + "_mean"),
			"RED": image.Select(redBand + "_mean"),
		})

	// Determine scale
	scale := defaultImageryScale
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Sample at the point
	result, err := savi.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		ComputeFloat(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to compute SAVI: %w", err)
	}

	return result, nil
}

// NDWI calculates the Normalized Difference Water Index at a point.
//
// NDWI = (Green - NIR) / (Green + NIR)
//
// NDWI is used to detect water bodies and monitor water content.
// Positive values indicate water presence.
//
// Example:
//
//	ndwi, err := helpers.NDWI(client, 45.5152, -122.6784, "2023-06-01")
func NDWI(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	ctx := context.Background()
	return NDWIWithContext(ctx, client, lat, lon, date, opts...)
}

// NDWIWithContext is like NDWI but accepts a context.
func NDWIWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the appropriate band names
	greenBand, nirBand := getBandNamesForWater(cfg.dataset)

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

	// Select Green and NIR bands, calculate NDWI using normalized difference
	// NDWI = (Green - NIR) / (Green + NIR)
	image := collection.
		Select(greenBand, nirBand).
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
		return 0, fmt.Errorf("failed to compute NDWI: %w", err)
	}

	return result, nil
}

// NDBI calculates the Normalized Difference Built-up Index at a point.
//
// NDBI = (SWIR - NIR) / (SWIR + NIR)
//
// NDBI is used to detect built-up (urban) areas.
// Positive values indicate built-up areas.
//
// Example:
//
//	ndbi, err := helpers.NDBI(client, 45.5152, -122.6784, "2023-06-01")
func NDBI(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	ctx := context.Background()
	return NDBIWithContext(ctx, client, lat, lon, date, opts...)
}

// NDBIWithContext is like NDBI but accepts a context.
func NDBIWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return 0, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get the appropriate band names
	swirBand, nirBand := getBandNamesForBuiltUp(cfg.dataset)

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

	// Select SWIR and NIR bands, calculate NDBI using normalized difference
	// NDBI = (SWIR - NIR) / (SWIR + NIR)
	image := collection.
		Select(swirBand, nirBand).
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
		return 0, fmt.Errorf("failed to compute NDBI: %w", err)
	}

	return result, nil
}

// SpectralBands returns the spectral band values at a point.
//
// Returns a map of band names to reflectance values.
// Band names depend on the satellite:
//   - Landsat: B1 (Coastal), B2 (Blue), B3 (Green), B4 (Red), B5 (NIR), B6 (SWIR1), B7 (SWIR2)
//   - Sentinel-2: B1-B12
//
// Example:
//
//	bands, err := helpers.SpectralBands(client, 45.5152, -122.6784, "2023-06-01",
//	    helpers.Landsat8())
//	fmt.Printf("Red: %.4f, NIR: %.4f\n", bands["B4"], bands["B5"])
func SpectralBands(client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (map[string]float64, error) {
	ctx := context.Background()
	return SpectralBandsWithContext(ctx, client, lat, lon, date, opts...)
}

// SpectralBandsWithContext is like SpectralBands but accepts a context.
func SpectralBandsWithContext(ctx context.Context, client *earthengine.Client, lat, lon float64, date string, opts ...ImageryOption) (map[string]float64, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	// Apply options
	cfg := &imageryConfig{
		dataset: landsat8DatasetID, // Default to Landsat 8
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Get bands to sample based on dataset
	var bands []string
	switch cfg.dataset {
	case landsat8DatasetID, landsat9DatasetID:
		// Landsat 8/9 surface reflectance bands
		bands = []string{"SR_B1", "SR_B2", "SR_B3", "SR_B4", "SR_B5", "SR_B6", "SR_B7"}
	case sentinel2DatasetID:
		// Sentinel-2 bands (10m, 20m, 60m)
		bands = []string{"B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B8A", "B9", "B11", "B12"}
	case modisVIDatasetID:
		// MODIS surface reflectance bands
		bands = []string{"sur_refl_b01", "sur_refl_b02", "sur_refl_b03", "sur_refl_b04", "sur_refl_b05", "sur_refl_b06", "sur_refl_b07"}
	default:
		return nil, fmt.Errorf("unsupported dataset for spectral bands: %s", cfg.dataset)
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

	// Get mean image
	image := collection.Select(bands...).Reduce(earthengine.ReducerMean())

	// Determine scale
	scale := defaultImageryScale
	if cfg.scale != nil {
		scale = *cfg.scale
	}

	// Sample at the point - this will return a dictionary with all band values
	result, err := image.
		ReduceRegion(
			earthengine.NewPoint(lon, lat),
			earthengine.ReducerFirst(),
			earthengine.Scale(scale),
		).
		Compute(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to compute spectral bands: %w", err)
	}

	// Convert to map[string]float64
	bandValues := make(map[string]float64)
	for key, value := range result {
		if floatVal, ok := value.(float64); ok {
			// Remove "_mean" suffix added by Reduce
			bandName := key
			if len(key) > 5 && key[len(key)-5:] == "_mean" {
				bandName = key[:len(key)-5]
			}
			bandValues[bandName] = floatVal
		}
	}

	return bandValues, nil
}

// CompositeMethod represents different compositing methods.
type CompositeMethod string

const (
	// MedianComposite takes the median value across time series.
	MedianComposite CompositeMethod = "median"
	// MeanComposite takes the mean value across time series.
	MeanComposite CompositeMethod = "mean"
	// MosaicComposite creates a mosaic (most recent on top).
	MosaicComposite CompositeMethod = "mosaic"
	// GreenestPixelComposite selects the greenest pixel (highest NDVI).
	GreenestPixelComposite CompositeMethod = "greenest"
)

// Composite creates a composite image from a time series.
//
// A composite combines multiple images over a time period into a single image.
// This is useful for creating cloud-free images or seasonal summaries.
//
// Note: This is a placeholder. Requires ImageCollection filtering and reduction.
//
// Example:
//
//	// Create a summer 2023 median composite
//	composite, err := helpers.Composite(client, bounds,
//	    "2023-06-01", "2023-08-31",
//	    helpers.Sentinel2(),
//	    helpers.CloudMask(20))
func Composite(client *earthengine.Client, bounds Bounds, startDate, endDate string, method CompositeMethod, opts ...ImageryOption) error {
	ctx := context.Background()
	return CompositeWithContext(ctx, client, bounds, startDate, endDate, method, opts...)
}

// CompositeWithContext is like Composite but accepts a context.
func CompositeWithContext(ctx context.Context, client *earthengine.Client, bounds Bounds, startDate, endDate string, method CompositeMethod, opts ...ImageryOption) error {
	if err := bounds.Validate(); err != nil {
		return err
	}

	// Placeholder - requires ImageCollection filtering and reduction
	return fmt.Errorf("Composite creation requires ImageCollection filtering support (not yet implemented)")
}

// NDVIQuery represents a deferred NDVI query for batch operations.
type NDVIQuery struct {
	lat  float64
	lon  float64
	date string
	opts []ImageryOption
}

// NewNDVIQuery creates a new NDVI query for batch execution.
func NewNDVIQuery(lat, lon float64, date string, opts ...ImageryOption) Query {
	return &NDVIQuery{
		lat:  lat,
		lon:  lon,
		date: date,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *NDVIQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return NDVIWithContext(ctx, client, q.lat, q.lon, q.date, q.opts...)
}

// EVIQuery represents a deferred EVI query for batch operations.
type EVIQuery struct {
	lat  float64
	lon  float64
	date string
	opts []ImageryOption
}

// NewEVIQuery creates a new EVI query for batch execution.
func NewEVIQuery(lat, lon float64, date string, opts ...ImageryOption) Query {
	return &EVIQuery{
		lat:  lat,
		lon:  lon,
		date: date,
		opts: opts,
	}
}

// Execute implements the Query interface.
func (q *EVIQuery) Execute(ctx context.Context, client *earthengine.Client) (interface{}, error) {
	return EVIWithContext(ctx, client, q.lat, q.lon, q.date, q.opts...)
}
