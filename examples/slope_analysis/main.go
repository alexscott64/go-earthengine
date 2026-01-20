package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/go-earthengine"
	"github.com/yourusername/go-earthengine/helpers"
)

// This example demonstrates how to analyze terrain slope for various use cases.
//
// Slope analysis is useful for:
// - Construction site evaluation
// - Road planning and engineering
// - Avalanche and landslide risk assessment
// - Agricultural land suitability
// - Hiking trail difficulty classification
// - Solar panel installation planning
// - Drainage and erosion analysis

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	client, err := earthengine.NewClient(ctx, "service-account.json")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Terrain Slope Analysis Examples")
	fmt.Println("================================")
	fmt.Println()

	// Example 1: Single location slope analysis
	example1_SingleLocation(ctx, client)

	// Example 2: Batch analysis for multiple sites
	example2_BatchAnalysis(ctx, client)

	// Example 3: Construction site evaluation
	example3_ConstructionSite(ctx, client)

	// Example 4: Hiking trail difficulty
	example4_HikingTrail(ctx, client)

	// Example 5: Agricultural suitability
	example5_Agricultural(ctx, client)
}

func example1_SingleLocation(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Single Location Analysis")
	fmt.Println("------------------------------------")

	// Mount Rainier summit area
	lat, lon := 46.8523, -121.7603

	// Get slope
	slope, err := helpers.Slope(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Get aspect (direction slope faces)
	aspect, err := helpers.Aspect(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Get elevation for context
	elevation, err := helpers.Elevation(client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Location: Mount Rainier summit area (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("  Elevation: %.0f meters\n", elevation)
	fmt.Printf("  Slope: %.1f degrees\n", slope)
	fmt.Printf("  Aspect: %.0f degrees (%s)\n", aspect, aspectToDirection(aspect))
	fmt.Printf("  Classification: %s\n", classifySlope(slope))
	fmt.Println()
}

func example2_BatchAnalysis(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Batch Analysis for Multiple Sites")
	fmt.Println("---------------------------------------------")

	locations := []struct {
		name     string
		lat, lon float64
		purpose  string
	}{
		{"Flat farmland", 41.2565, -95.9345, "Agriculture"},
		{"Moderate hill", 37.8651, -119.5383, "Residential"},
		{"Steep mountain", 46.8523, -121.7603, "Recreation"},
		{"Coastal plain", 47.6062, -122.3321, "Urban development"},
	}

	batch := helpers.NewBatch(client, 10)
	for _, loc := range locations {
		batch.Add(helpers.NewSlopeQuery(loc.lat, loc.lon, helpers.SRTM()))
	}

	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Batch error: %v", err)
		return
	}

	for i, result := range results {
		if result.Error != nil {
			log.Printf("Error for %s: %v", locations[i].name, result.Error)
			continue
		}

		slope := result.Value.(float64)
		suitability := evaluateSuitability(slope, locations[i].purpose)

		fmt.Printf("%s (%.4f, %.4f)\n", locations[i].name, locations[i].lat, locations[i].lon)
		fmt.Printf("  Purpose: %s\n", locations[i].purpose)
		fmt.Printf("  Slope: %.1f degrees\n", slope)
		fmt.Printf("  Suitability: %s\n", suitability)
		fmt.Println()
	}
}

func example3_ConstructionSite(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Construction Site Evaluation")
	fmt.Println("---------------------------------------")

	// Potential construction site
	lat, lon := 47.6500, -122.1200

	slope, err := helpers.Slope(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	aspect, err := helpers.Aspect(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Construction Site: (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("  Slope: %.1f degrees\n", slope)
	fmt.Printf("  Aspect: %.0f degrees (%s)\n", aspect, aspectToDirection(aspect))
	fmt.Println()
	fmt.Println("  Construction Assessment:")

	// Evaluate different aspects
	if slope < 2 {
		fmt.Println("    ✓ Excellent - Minimal grading required")
	} else if slope < 5 {
		fmt.Println("    ✓ Good - Minor grading needed")
	} else if slope < 10 {
		fmt.Println("    ⚠ Moderate - Significant grading required")
	} else if slope < 15 {
		fmt.Println("    ⚠ Challenging - Major earthwork needed")
	} else {
		fmt.Println("    ✗ Not recommended - Excessive slope")
	}

	// Drainage considerations
	if slope >= 1 && slope <= 3 {
		fmt.Println("    ✓ Good drainage potential")
	} else if slope < 1 {
		fmt.Println("    ⚠ May require drainage improvements")
	} else {
		fmt.Println("    ⚠ Erosion control measures required")
	}

	fmt.Println()
}

func example4_HikingTrail(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Hiking Trail Difficulty Assessment")
	fmt.Println("---------------------------------------------")

	// Points along a hiking trail
	trailPoints := []struct {
		name     string
		lat, lon float64
	}{
		{"Trailhead", 47.4400, -121.7900},
		{"Mid-point", 47.4450, -121.7950},
		{"Summit", 47.4500, -121.8000},
	}

	fmt.Println("Trail segment analysis:")
	fmt.Println()

	totalDistance := 0.0
	for i, point := range trailPoints {
		slope, err := helpers.Slope(ctx, client, point.lat, point.lon, helpers.SRTM())
		if err != nil {
			log.Printf("Error at %s: %v", point.name, err)
			continue
		}

		elevation, err := helpers.Elevation(client, point.lat, point.lon, helpers.SRTM())
		if err != nil {
			log.Printf("Error getting elevation: %v", err)
			continue
		}

		difficulty := classifyHikingDifficulty(slope)
		fmt.Printf("%d. %s (%.4f, %.4f)\n", i+1, point.name, point.lat, point.lon)
		fmt.Printf("   Elevation: %.0f m\n", elevation)
		fmt.Printf("   Slope: %.1f degrees\n", slope)
		fmt.Printf("   Difficulty: %s\n", difficulty)
		fmt.Println()
	}
}

func example5_Agricultural(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Agricultural Land Suitability")
	fmt.Println("----------------------------------------")

	// Potential farmland
	lat, lon := 41.3000, -96.0000

	slope, err := helpers.Slope(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	aspect, err := helpers.Aspect(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Agricultural Site: (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("  Slope: %.1f degrees\n", slope)
	fmt.Printf("  Aspect: %.0f degrees (%s)\n", aspect, aspectToDirection(aspect))
	fmt.Println()
	fmt.Println("  Crop Suitability:")

	// Different crops have different slope requirements
	evaluateCropSuitability("Row crops (corn, soybeans)", slope, 0, 3)
	evaluateCropSuitability("Small grains (wheat, oats)", slope, 0, 8)
	evaluateCropSuitability("Hay/Pasture", slope, 0, 15)
	evaluateCropSuitability("Orchards/Vineyards", slope, 3, 15)

	fmt.Println()
	fmt.Println("  Erosion Risk:")
	if slope < 2 {
		fmt.Println("    Low - Minimal erosion concerns")
	} else if slope < 5 {
		fmt.Println("    Moderate - Basic conservation practices recommended")
	} else if slope < 10 {
		fmt.Println("    High - Contour farming and terracing recommended")
	} else {
		fmt.Println("    Severe - Not recommended for cultivation")
	}

	fmt.Println()
}

// Helper functions

func classifySlope(slope float64) string {
	switch {
	case slope < 2:
		return "Nearly level (0-2°)"
	case slope < 5:
		return "Gently sloping (2-5°)"
	case slope < 10:
		return "Moderately sloping (5-10°)"
	case slope < 15:
		return "Strongly sloping (10-15°)"
	case slope < 25:
		return "Steep (15-25°)"
	case slope < 35:
		return "Very steep (25-35°)"
	default:
		return "Extremely steep (>35°)"
	}
}

func aspectToDirection(aspect float64) string {
	switch {
	case aspect < 22.5 || aspect >= 337.5:
		return "North"
	case aspect < 67.5:
		return "Northeast"
	case aspect < 112.5:
		return "East"
	case aspect < 157.5:
		return "Southeast"
	case aspect < 202.5:
		return "South"
	case aspect < 247.5:
		return "Southwest"
	case aspect < 292.5:
		return "West"
	default:
		return "Northwest"
	}
}

func evaluateSuitability(slope float64, purpose string) string {
	switch purpose {
	case "Agriculture":
		if slope < 3 {
			return "Excellent"
		} else if slope < 8 {
			return "Good"
		} else if slope < 15 {
			return "Marginal"
		}
		return "Not suitable"

	case "Residential":
		if slope < 5 {
			return "Excellent"
		} else if slope < 10 {
			return "Good"
		} else if slope < 15 {
			return "Fair"
		}
		return "Challenging"

	case "Recreation":
		if slope >= 15 && slope <= 35 {
			return "Excellent for hiking/skiing"
		} else if slope < 15 {
			return "Good for trails"
		}
		return "Expert terrain"

	case "Urban development":
		if slope < 2 {
			return "Ideal"
		} else if slope < 5 {
			return "Good"
		} else if slope < 10 {
			return "Moderate cost"
		}
		return "High development cost"
	}

	return "Unknown"
}

func classifyHikingDifficulty(slope float64) string {
	switch {
	case slope < 5:
		return "Easy"
	case slope < 10:
		return "Moderate"
	case slope < 15:
		return "Moderately Difficult"
	case slope < 25:
		return "Difficult"
	default:
		return "Very Difficult"
	}
}

func evaluateCropSuitability(cropName string, slope, minSlope, maxSlope float64) {
	if slope >= minSlope && slope <= maxSlope {
		fmt.Printf("    ✓ %s: Suitable\n", cropName)
	} else if slope < minSlope {
		fmt.Printf("    ⚠ %s: Too flat (drainage issues)\n", cropName)
	} else {
		fmt.Printf("    ✗ %s: Too steep (erosion risk)\n", cropName)
	}
}
