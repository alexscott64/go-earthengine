package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexscott64/go-earthengine"
	"github.com/alexscott64/go-earthengine/helpers"
)

// This example demonstrates advanced compositing methods for creating
// high-quality cloud-free imagery.
//
// Use cases:
// - Creating cloud-free base maps
// - Vegetation phenology studies
// - Land cover classification
// - Change detection
// - Time-lapse animations

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	// Note: This example requires proper authentication setup
	client := &earthengine.Client{}
	_ = ctx // For demonstration only

	fmt.Println("Advanced Compositing Examples")
	fmt.Println("==============================")
	fmt.Println()

	// Example 1: Quality mosaic
	example1_QualityMosaic(ctx, client)

	// Example 2: Greenest pixel composite
	example2_GreenestPixel(ctx, client)

	// Example 3: Percentile composites
	example3_Percentiles(ctx, client)

	// Example 4: Seasonal composites
	example4_Seasonal(ctx, client)

	// Example 5: Temporal smoothing
	example5_TemporalSmoothing(ctx, client)

	// Example 6: Composite comparison
	example6_CompositeComparison(ctx, client)
}

func example1_QualityMosaic(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Quality Mosaic")
	fmt.Println("-------------------------")

	// Create image collection for summer 2023
	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate("2023-06-01", "2023-08-31").
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 30)

	fmt.Println("Dataset: Sentinel-2")
	fmt.Println("Period: Summer 2023")
	fmt.Println("Method: Quality mosaic (selects best pixels)")
	fmt.Println()

	// Create quality mosaic using cloud probability as quality metric
	mosaic, err := helpers.QualityMosaic(ctx, client, collection,
		"MSK_CLDPRB") // Cloud probability band (lower is better)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("✓ Quality mosaic created")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  - Each pixel is selected from the image with the best quality score")
	fmt.Println("  - Uses cloud probability to prioritize clear pixels")
	fmt.Println("  - Results in a cloud-free composite with highest quality pixels")
	fmt.Println()

	_ = mosaic
}

func example2_GreenestPixel(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Greenest Pixel Composite")
	fmt.Println("------------------------------------")

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate("2023-01-01", "2023-12-31").
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	fmt.Println("Dataset: Sentinel-2")
	fmt.Println("Period: Full year 2023")
	fmt.Println("Method: Greenest pixel (maximum NDVI)")
	fmt.Println()

	// Create greenest pixel composite
	greenest, err := helpers.CreateGreenestPixelComposite(ctx, client, collection)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("✓ Greenest pixel composite created")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  - For each pixel, selects the observation with highest NDVI")
	fmt.Println("  - Captures peak greenness/vegetation health")
	fmt.Println("  - Excellent for vegetation mapping and agricultural monitoring")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Peak growing season imagery")
	fmt.Println("  - Forest health assessment")
	fmt.Println("  - Crop phenology studies")

	fmt.Println()
	_ = greenest
}

func example3_Percentiles(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Percentile Composites")
	fmt.Println("---------------------------------")

	collection := client.ImageCollection("LANDSAT/LC08/C02/T1_L2").
		FilterDate("2023-01-01", "2023-12-31").
		FilterMetadata("CLOUD_COVER", "less_than", 30)

	fmt.Println("Dataset: Landsat 8")
	fmt.Println("Period: Full year 2023")
	fmt.Println()

	percentiles := []float64{10, 50, 90}

	for _, p := range percentiles {
		composite, err := helpers.CreatePercentileComposite(ctx, client, collection, p)
		if err != nil {
			log.Printf("Error creating %gth percentile: %v", p, err)
			continue
		}

		fmt.Printf("✓ %gth percentile composite created\n", p)
		_ = composite
	}

	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  10th percentile: Dark pixels (shadows, water)")
	fmt.Println("  50th percentile: Median values (typical conditions)")
	fmt.Println("  90th percentile: Bright pixels (clouds, snow)")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - 10th: Water mapping, shadow analysis")
	fmt.Println("  - 50th: Cloud-free imagery, typical conditions")
	fmt.Println("  - 90th: Snow mapping, cloud detection")

	fmt.Println()
}

func example4_Seasonal(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Seasonal Composites")
	fmt.Println("------------------------------")

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	fmt.Println("Dataset: Sentinel-2")
	fmt.Println("Year: 2023")
	fmt.Println("Method: Median composite per season")
	fmt.Println()

	// Create seasonal composites
	seasons, err := helpers.SeasonalComposite(ctx, client, collection, 2023)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Display results
	seasonNames := []string{"spring", "summer", "fall", "winter"}
	for _, season := range seasonNames {
		if composite, exists := seasons[season]; exists {
			fmt.Printf("✓ %s composite created\n", season)
			_ = composite
		}
	}

	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  - Creates separate composites for each season")
	fmt.Println("  - Spring: March-May")
	fmt.Println("  - Summer: June-August")
	fmt.Println("  - Fall: September-November")
	fmt.Println("  - Winter: December-February")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Phenology studies")
	fmt.Println("  - Seasonal change detection")
	fmt.Println("  - Agricultural monitoring")
	fmt.Println("  - Climate analysis")

	fmt.Println()
}

func example5_TemporalSmoothing(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Temporal Smoothing")
	fmt.Println("------------------------------")

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate("2023-01-01", "2023-12-31").
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	fmt.Println("Dataset: Sentinel-2")
	fmt.Println("Period: Full year 2023")
	fmt.Println("Method: Moving window smoothing")
	fmt.Println()

	// Apply temporal smoothing with 5-image window
	smoothed, err := helpers.TemporalSmoothingComposite(ctx, client,
		collection, 5)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("✓ Created %d smoothed images\n", len(smoothed))
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  - Applies moving average filter to image time series")
	fmt.Println("  - Reduces noise and artifacts")
	fmt.Println("  - Preserves temporal trends")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - NDVI time series analysis")
	fmt.Println("  - Phenology curve fitting")
	fmt.Println("  - Noise reduction")
	fmt.Println("  - Gap filling")

	fmt.Println()
}

func example6_CompositeComparison(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 6: Comparing Composites")
	fmt.Println("--------------------------------")

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate("2023-06-01", "2023-08-31").
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	fmt.Println("Dataset: Sentinel-2")
	fmt.Println("Period: Summer 2023")
	fmt.Println()

	// Create median composite
	medianConfig := helpers.CompositeConfig{
		Method:         helpers.MedianComposite,
		CloudThreshold: 20,
		Scale:          10,
	}
	median, err := helpers.AdvancedComposite(ctx, client, collection, medianConfig)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Create greenest pixel composite
	greenestConfig := helpers.CompositeConfig{
		Method:         helpers.GreenestPixelComposite,
		CloudThreshold: 20,
		Scale:          10,
	}
	greenest, err := helpers.AdvancedComposite(ctx, client, collection, greenestConfig)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Comparing Methods:")
	fmt.Println()

	fmt.Println("Median Composite:")
	fmt.Printf("  Images used: %d\n", median.ObservationCount)
	fmt.Printf("  Method: %s\n", median.Method)
	fmt.Println("  Pros: Reduces noise, handles outliers")
	fmt.Println("  Cons: May miss peak values")
	fmt.Println()

	fmt.Println("Greenest Pixel:")
	fmt.Printf("  Images used: %d\n", greenest.ObservationCount)
	fmt.Printf("  Method: %s\n", greenest.Method)
	fmt.Println("  Pros: Captures peak greenness")
	fmt.Println("  Cons: May select cloudy pixels if NDVI is high")
	fmt.Println()

	// Calculate composite metrics
	medianMetrics, _ := helpers.CalculateCompositeMetrics(ctx, client, median.Image)
	greenestMetrics, _ := helpers.CalculateCompositeMetrics(ctx, client, greenest.Image)

	fmt.Println("Quality Metrics:")
	fmt.Println()
	fmt.Printf("Median - Coverage: %.1f%%, Cloud-free: %.1f%%\n",
		medianMetrics.Coverage*100,
		medianMetrics.CloudFreePixels*100)
	fmt.Printf("Greenest - Coverage: %.1f%%, Cloud-free: %.1f%%\n",
		greenestMetrics.Coverage*100,
		greenestMetrics.CloudFreePixels*100)

	fmt.Println()

	// Compare the composites
	diff, err := helpers.CompareComposites(ctx, client, median.Image,
		greenest.Image, "NDVI")
	if err == nil {
		fmt.Println("Difference Analysis (NDVI band):")
		fmt.Printf("  Mean difference: %.3f\n", diff.MeanDifference)
		fmt.Printf("  Pixels changed: %.1f%%\n", diff.PercentChanged)
		fmt.Printf("  Max difference: %.3f\n", diff.MaxDifference)
	}

	fmt.Println()
}

// Additional helper functions for real-world scenarios

func createAnnualComposite(ctx context.Context, client *earthengine.Client, year int, method helpers.CompositeMethod) (*helpers.CompositeResult, error) {
	// Create annual composite with specified method
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate(startDate, endDate).
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	config := helpers.CompositeConfig{
		Method:         method,
		CloudThreshold: 20,
		Scale:          10,
	}

	return helpers.AdvancedComposite(ctx, client, collection, config)
}

func compareCompositeMethodsForRegion(ctx context.Context, client *earthengine.Client, region *earthengine.Geometry, startDate, endDate string) (map[string]*helpers.CompositeResult, error) {
	// Compare different compositing methods for a region
	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate(startDate, endDate)

	methods := []helpers.CompositeMethod{
		helpers.MedianComposite,
		helpers.MeanComposite,
		helpers.GreenestPixelComposite,
	}

	results := make(map[string]*helpers.CompositeResult)

	for _, method := range methods {
		config := helpers.CompositeConfig{
			Method:         method,
			CloudThreshold: 20,
			Region:         region,
		}

		result, err := helpers.AdvancedComposite(ctx, client, collection, config)
		if err != nil {
			continue
		}

		results[string(method)] = result
	}

	return results, nil
}

func createMultiYearComposite(ctx context.Context, client *earthengine.Client, startYear, endYear int) ([]*helpers.CompositeResult, error) {
	// Create annual composites for multiple years
	results := make([]*helpers.CompositeResult, 0)

	for year := startYear; year <= endYear; year++ {
		composite, err := createAnnualComposite(ctx, client, year, helpers.MedianComposite)
		if err != nil {
			log.Printf("Warning: Failed to create composite for year %d: %v", year, err)
			continue
		}

		results = append(results, composite)
	}

	return results, nil
}
