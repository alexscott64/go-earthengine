package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ee "github.com/yourusername/go-earthengine"
	"github.com/yourusername/go-earthengine/helpers"
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

	// Example 1: Simple tree coverage query using helpers
	fmt.Println("=== Example 1: Simple Tree Coverage Query ===")
	latitude := 47.6
	longitude := -120.9

	coverage, err := helpers.TreeCoverage(client, latitude, longitude)
	if err != nil {
		log.Fatalf("Failed to get tree coverage: %v", err)
	}

	fmt.Printf("Location: (%.4f, %.4f)\n", latitude, longitude)
	fmt.Printf("Tree Canopy Coverage: %.2f%%\n\n", coverage)

	// Example 2: Tree coverage with options (specific year)
	fmt.Println("=== Example 2: Tree Coverage for Specific Year ===")
	coverage2020, err := helpers.TreeCoverage(client, latitude, longitude, helpers.Year(2020))
	if err != nil {
		log.Fatalf("Failed to get tree coverage: %v", err)
	}

	fmt.Printf("Tree Coverage in 2020: %.2f%%\n\n", coverage2020)

	// Example 3: Multiple locations using batch processing
	fmt.Println("=== Example 3: Batch Processing Multiple Locations ===")
	locations := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"Seattle, WA", 47.6062, -122.3321},
		{"Portland, OR", 45.5152, -122.6784},
		{"Spokane, WA", 47.6588, -117.4260},
	}

	// Create batch with 3 concurrent queries
	batch := helpers.NewBatch(client, 3)
	for _, loc := range locations {
		batch.Add(helpers.NewTreeCoverageQuery(loc.latitude, loc.longitude))
	}

	results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
		fmt.Printf("Progress: %d/%d\n", completed, total)
	})
	if err != nil {
		log.Fatalf("Batch execution failed: %v", err)
	}

	// Display results
	for i, result := range results {
		if result.Error != nil {
			log.Printf("Error for %s: %v", locations[i].name, result.Error)
			continue
		}
		coverage := result.Value.(float64)
		fmt.Printf("%s: %.2f%% tree coverage\n", locations[i].name, coverage)
	}

	// Example 4: Other helper functions
	fmt.Println("\n=== Example 4: Other Helper Functions ===")

	// Get elevation
	elevation, err := helpers.Elevation(client, latitude, longitude)
	if err != nil {
		log.Printf("Failed to get elevation: %v", err)
	} else {
		fmt.Printf("Elevation: %.0f meters\n", elevation)
	}

	// Check if urban
	urban, err := helpers.IsUrban(client, latitude, longitude)
	if err != nil {
		log.Printf("Failed to check urban status: %v", err)
	} else {
		if urban {
			fmt.Println("Location: Urban area")
		} else {
			fmt.Println("Location: Non-urban area")
		}
	}
}
