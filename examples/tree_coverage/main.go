package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ee "github.com/yourusername/go-earthengine"
)

func main() {
	ctx := context.Background()

	// Create client using environment variables from .env file
	client, err := ee.NewClient(ctx,
		ee.WithServiceAccountEnv(),
		ee.WithProject(os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")),
	)
	if err != nil {
		log.Fatalf("Failed to create Earth Engine client: %v", err)
	}

	// Example 1: Simple tree coverage query
	fmt.Println("=== Example 1: Simple Tree Coverage Query ===")
	latitude := 47.6
	longitude := -120.9

	coverage, err := client.GetTreeCoverage(ctx, latitude, longitude)
	if err != nil {
		log.Fatalf("Failed to get tree coverage: %v", err)
	}

	fmt.Printf("Location: (%.4f, %.4f)\n", latitude, longitude)
	fmt.Printf("Tree Canopy Coverage: %.2f%%\n\n", coverage)

	// Example 2: Detailed tree coverage with metadata
	fmt.Println("=== Example 2: Detailed Tree Coverage ===")
	result, err := client.GetTreeCoverageDetailed(ctx, latitude, longitude)
	if err != nil {
		log.Fatalf("Failed to get detailed tree coverage: %v", err)
	}

	fmt.Printf("Latitude: %.4f\n", result.Latitude)
	fmt.Printf("Longitude: %.4f\n", result.Longitude)
	fmt.Printf("Coverage: %.2f%%\n", result.Coverage)
	fmt.Printf("Data Source: %s\n\n", result.DataSource)

	// Example 3: Multiple locations
	fmt.Println("=== Example 3: Multiple Locations ===")
	locations := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"Seattle, WA", 47.6062, -122.3321},
		{"Portland, OR", 45.5152, -122.6784},
		{"Spokane, WA", 47.6588, -117.4260},
	}

	for _, loc := range locations {
		coverage, err := client.GetTreeCoverage(ctx, loc.latitude, loc.longitude)
		if err != nil {
			log.Printf("Error for %s: %v", loc.name, err)
			continue
		}
		fmt.Printf("%s: %.2f%% tree coverage\n", loc.name, coverage)
	}
}
