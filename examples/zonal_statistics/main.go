package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alexscott64/go-earthengine"
	"github.com/alexscott64/go-earthengine/helpers"
)

// This example demonstrates zonal statistics for analyzing image values
// within polygon boundaries.
//
// Use cases:
// - Agricultural field statistics
// - Watershed analysis
// - Administrative region summaries
// - Protected area monitoring
// - Urban heat island analysis

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	// Note: This example requires proper authentication setup
	client := &earthengine.Client{}
	_ = ctx // For demonstration only

	fmt.Println("Zonal Statistics Examples")
	fmt.Println("=========================")
	fmt.Println()

	// Example 1: Basic zonal statistics
	example1_BasicZonalStats(ctx, client)

	// Example 2: Multiple statistics
	example2_MultipleStatistics(ctx, client)

	// Example 3: Agricultural field analysis
	example3_AgFieldAnalysis(ctx, client)

	// Example 4: Watershed comparison
	example4_WatershedComparison(ctx, client)

	// Example 5: Time series by zone
	example5_ZonalTimeSeries(ctx, client)

	// Example 6: Change detection by zone
	example6_ZonalChangeDetection(ctx, client)

	// Example 7: Export results
	example7_ExportResults(ctx, client)
}

func example1_BasicZonalStats(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Basic Zonal Statistics")
	fmt.Println("----------------------------------")

	// Create sample polygons (watersheds, admin boundaries, etc.)
	zones := &earthengine.FeatureCollection{}

	// Load NDVI image
	image := client.Image("COPERNICUS/S2_SR/20230701T...")

	fmt.Println("Dataset: Sentinel-2 NDVI")
	fmt.Println("Date: 2023-07-01")
	fmt.Println("Zones: 5 watersheds")
	fmt.Println()

	// Calculate mean NDVI per zone
	result, err := helpers.ZonalMean(ctx, client, image, zones, 10)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Mean NDVI by Zone:")
	for _, zone := range result.Zones {
		if ndvi, exists := zone.Stats["NDVI_mean"]; exists {
			fmt.Printf("  Zone %v: %.3f (%.0f pixels, %.1f ha)\n",
				zone.ZoneID,
				ndvi,
				float64(zone.PixelCount),
				zone.Area/10000) // Convert m² to hectares
		}
	}

	fmt.Println()
	_ = result
}

func example2_MultipleStatistics(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Multiple Statistics")
	fmt.Println("-------------------------------")

	zones := &earthengine.FeatureCollection{}
	image := client.Image("COPERNICUS/S2_SR/20230701T...")

	fmt.Println("Dataset: Sentinel-2 Surface Reflectance")
	fmt.Println("Bands: B4 (Red), B8 (NIR)")
	fmt.Println()

	// Calculate multiple statistics
	config := helpers.ZonalStatsConfig{
		Statistics: []helpers.ZonalStatistic{
			helpers.Mean,
			helpers.StdDev,
			helpers.Min,
			helpers.Max,
		},
		Scale:     10,
		Bands:     []string{"B4", "B8"},
		ZoneIDKey: "id",
	}

	result, err := helpers.CalculateZonalStats(ctx, client, image, zones, config)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Statistics by Zone:")
	for _, zone := range result.Zones {
		fmt.Printf("\nZone %v:\n", zone.ZoneID)
		fmt.Println("  Band | Mean  | StdDev | Min   | Max")
		fmt.Println("  -----|-------|--------|-------|-------")

		for _, band := range result.Bands {
			meanKey := fmt.Sprintf("%s_mean", band)
			stdKey := fmt.Sprintf("%s_stdDev", band)
			minKey := fmt.Sprintf("%s_min", band)
			maxKey := fmt.Sprintf("%s_max", band)

			fmt.Printf("  %-4s | %.3f | %.3f  | %.3f | %.3f\n",
				band,
				zone.Stats[meanKey],
				zone.Stats[stdKey],
				zone.Stats[minKey],
				zone.Stats[maxKey])
		}
	}

	fmt.Println()
}

func example3_AgFieldAnalysis(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Agricultural Field Analysis")
	fmt.Println("---------------------------------------")

	// Field boundaries
	fields := &earthengine.FeatureCollection{}

	// NDVI image
	image := client.Image("COPERNICUS/S2_SR/20230701T...")

	fmt.Println("Application: Crop health monitoring")
	fmt.Println("Fields: 10 agricultural parcels")
	fmt.Println("Metric: NDVI (vegetation health)")
	fmt.Println()

	// Calculate stats per field
	config := helpers.ZonalStatsConfig{
		Statistics: []helpers.ZonalStatistic{
			helpers.Mean,
			helpers.StdDev,
		},
		Scale:     10,
		Bands:     []string{"NDVI"},
		ZoneIDKey: "field_id",
	}

	result, err := helpers.CalculateZonalStats(ctx, client, image, fields, config)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Field Health Assessment:")
	fmt.Println("Field | Mean NDVI | StdDev | Area (ha) | Status")
	fmt.Println("------|-----------|--------|-----------|-------------")

	for _, zone := range result.Zones {
		mean := zone.Stats["NDVI_mean"]
		stddev := zone.Stats["NDVI_stdDev"]
		area := zone.Area / 10000

		// Classify health
		status := "Healthy"
		if mean < 0.4 {
			status = "Stressed"
		} else if mean < 0.6 {
			status = "Moderate"
		}

		fmt.Printf("%-5v | %.3f     | %.3f  | %.1f      | %s\n",
			zone.ZoneID, mean, stddev, area, status)
	}

	fmt.Println()

	// Summary
	summary := helpers.SummarizeZonalStats(result, "NDVI_mean")
	fmt.Println("Summary Statistics:")
	fmt.Printf("  Fields analyzed: %d\n", summary.TotalZones)
	fmt.Printf("  Mean NDVI: %.3f\n", summary.MeanValue)
	fmt.Printf("  Range: %.3f - %.3f\n", summary.MinValue, summary.MaxValue)
	fmt.Printf("  Total area: %.1f ha\n", summary.TotalArea/10000)

	fmt.Println()
	_ = fields
}

func example4_WatershedComparison(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Watershed Comparison")
	fmt.Println("--------------------------------")

	watersheds := &earthengine.FeatureCollection{}
	landCover := client.Image("USGS/NLCD_RELEASES/2023_REL/NLCD")

	fmt.Println("Application: Land cover analysis by watershed")
	fmt.Println("Watersheds: 3 drainage basins")
	fmt.Println()

	// Calculate frequency of land cover classes
	freq, err := helpers.ZonalFrequencyTable(ctx, client, landCover,
		watersheds, "landcover", 30)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Land Cover Distribution:")
	for i, zf := range freq {
		fmt.Printf("\nWatershed %d:\n", i+1)
		fmt.Printf("  Dominant class: %d\n", zf.Dominant)
		fmt.Printf("  Diversity index: %.2f\n", zf.Diversity)
		fmt.Println("  Class distribution:")

		// Show top 3 classes
		count := 0
		for class, pixels := range zf.ClassCounts {
			if count >= 3 {
				break
			}
			pct := float64(pixels) * 100 / float64(getTotalPixels(zf.ClassCounts))
			fmt.Printf("    Class %d: %.1f%% (%d pixels)\n", class, pct, pixels)
			count++
		}
	}

	fmt.Println()
	_ = watersheds
}

func example5_ZonalTimeSeries(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Zonal Time Series")
	fmt.Println("-----------------------------")

	zones := &earthengine.FeatureCollection{}
	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate("2023-01-01", "2023-12-31")

	fmt.Println("Application: Vegetation phenology by region")
	fmt.Println("Period: 2023 (monthly)")
	fmt.Println("Metric: Mean NDVI")
	fmt.Println()

	// Get time series for each zone
	series, err := helpers.CalculateZonalTimeSeries(ctx, client, collection, zones,
		helpers.Mean, "NDVI", 10)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("NDVI Time Series:")
	for _, ts := range series {
		fmt.Printf("\nZone %v:\n", ts.ZoneID)
		fmt.Println("  Date       | NDVI")
		fmt.Println("  -----------|------")

		// Show first few points
		for i, point := range ts.Series {
			if i >= 6 { // Show first 6 months
				break
			}
			fmt.Printf("  %s | %.3f\n", point.Time, point.Value)
		}
	}

	fmt.Println()
	_ = zones
}

func example6_ZonalChangeDetection(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 6: Zonal Change Detection")
	fmt.Println("----------------------------------")

	zones := &earthengine.FeatureCollection{}
	image2020 := client.Image("COPERNICUS/S2_SR/20200701T...")
	image2023 := client.Image("COPERNICUS/S2_SR/20230701T...")

	fmt.Println("Application: Forest change analysis")
	fmt.Println("Period: 2020 vs 2023")
	fmt.Println("Zones: Protected areas")
	fmt.Println()

	// Compare before and after
	comparison, err := helpers.CalculateZonalComparison(ctx, client, image2020,
		image2023, zones, helpers.Mean, 10)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Change Analysis Results:")
	fmt.Println("Zone | 2020  | 2023  | Change | % Change | Status")
	fmt.Println("-----|-------|-------|--------|----------|-------------")

	for _, comp := range comparison {
		fmt.Printf("%-4v | %.3f | %.3f | %+.3f | %+6.1f%% | %s\n",
			comp.ZoneID,
			comp.BeforeValue,
			comp.AfterValue,
			comp.Difference,
			comp.PercentChange,
			comp.ChangeCategory)
	}

	fmt.Println()

	// Summary
	increases := 0
	decreases := 0
	stable := 0
	for _, comp := range comparison {
		switch comp.ChangeCategory {
		case "increase":
			increases++
		case "decrease":
			decreases++
		default:
			stable++
		}
	}

	fmt.Println("Summary:")
	fmt.Printf("  Zones with increase: %d\n", increases)
	fmt.Printf("  Zones with decrease: %d\n", decreases)
	fmt.Printf("  Zones stable: %d\n", stable)

	fmt.Println()
	_ = zones
}

func example7_ExportResults(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 7: Export Results")
	fmt.Println("-------------------------")

	zones := &earthengine.FeatureCollection{}
	image := client.Image("COPERNICUS/S2_SR/20230701T...")

	// Calculate statistics
	config := helpers.ZonalStatsConfig{
		Statistics: []helpers.ZonalStatistic{
			helpers.Mean,
			helpers.StdDev,
			helpers.Min,
			helpers.Max,
		},
		Scale: 10,
		Bands: []string{"NDVI", "B4", "B8"},
	}

	result, err := helpers.CalculateZonalStats(ctx, client, image, zones, config)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Exporting results...")
	fmt.Println()

	// Export to CSV
	csv, err := helpers.ExportZonalStatsToCSV(result)
	if err != nil {
		log.Printf("Error exporting CSV: %v", err)
		return
	}

	// Save to file
	filename := "zonal_stats.csv"
	err = os.WriteFile(filename, []byte(csv), 0644)
	if err != nil {
		log.Printf("Error writing file: %v", err)
		return
	}

	fmt.Printf("✓ Exported to %s\n", filename)
	fmt.Printf("  Zones: %d\n", len(result.Zones))
	fmt.Printf("  Statistics: %d\n", len(result.Statistics))
	fmt.Printf("  Bands: %d\n", len(result.Bands))
	fmt.Println()

	// Convert to feature collection
	fc, err := helpers.ZonalStatsToFeatureCollection(result)
	if err != nil {
		log.Printf("Error converting to FC: %v", err)
		return
	}

	fmt.Println("✓ Converted to FeatureCollection")
	fmt.Println("  Can be exported to GeoJSON, Shapefile, etc.")

	fmt.Println()
	_ = fc
}

// Helper functions

func getTotalPixels(classCounts map[int]int) int {
	total := 0
	for _, count := range classCounts {
		total += count
	}
	return total
}

// Additional real-world helper functions

func analyzeProtectedAreas(ctx context.Context, client *earthengine.Client, protectedAreas *earthengine.FeatureCollection, year int) (*helpers.ZonalStatsResult, error) {
	// Analyze vegetation health in protected areas
	startDate := fmt.Sprintf("%d-06-01", year)
	endDate := fmt.Sprintf("%d-08-31", year)

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate(startDate, endDate)

	// Get median composite
	composite, err := helpers.AdvancedComposite(ctx, client, collection,
		helpers.CompositeConfig{
			Method: helpers.MedianComposite,
		})
	if err != nil {
		return nil, err
	}

	// Calculate zonal statistics
	config := helpers.ZonalStatsConfig{
		Statistics: []helpers.ZonalStatistic{
			helpers.Mean,
			helpers.StdDev,
		},
		Scale:     10,
		Bands:     []string{"NDVI"},
		ZoneIDKey: "name",
	}

	return helpers.CalculateZonalStats(ctx, client, composite.Image,
		protectedAreas, config)
}

func compareWatershedHealth(ctx context.Context, client *earthengine.Client, watersheds *earthengine.FeatureCollection, year1, year2 int) ([]helpers.ZonalComparison, error) {
	// Compare watershed health between two years
	image1 := client.Image(fmt.Sprintf("COPERNICUS/S2_SR/%d0701T...", year1))
	image2 := client.Image(fmt.Sprintf("COPERNICUS/S2_SR/%d0701T...", year2))

	return helpers.CalculateZonalComparison(ctx, client, image1, image2,
		watersheds, helpers.Mean, 10)
}

func monitorAgFieldsThroughSeason(ctx context.Context, client *earthengine.Client, fields *earthengine.FeatureCollection, year int) ([]helpers.ZonalTimeSeries, error) {
	// Monitor agricultural fields through growing season
	startDate := fmt.Sprintf("%d-03-01", year)
	endDate := fmt.Sprintf("%d-10-31", year)

	collection := client.ImageCollection("COPERNICUS/S2_SR").
		FilterDate(startDate, endDate).
		FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)

	return helpers.CalculateZonalTimeSeries(ctx, client, collection, fields,
		helpers.Mean, "NDVI", 10)
}
