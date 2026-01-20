// +build ignore

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	ee "github.com/yourusername/go-earthengine"
	"github.com/yourusername/go-earthengine/helpers"
)

// loadEnv loads environment variables from a .env file
func loadEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"`)

		// Unescape newlines (double backslash)
		value = strings.ReplaceAll(value, `\\n`, "\n")

		os.Setenv(key, value)
	}

	return scanner.Err()
}

func main() {
	// Load .env file
	if err := loadEnv(".env"); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	ctx := context.Background()

	// Create client using environment variables
	client, err := ee.NewClient(ctx,
		ee.WithServiceAccountEnv(),
		ee.WithProject(os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")),
	)
	if err != nil {
		log.Fatalf("Failed to create Earth Engine client: %v", err)
	}

	fmt.Println("=== Testing Tree Coverage Query (Helpers) ===")

	// Test location in Washington state
	latitude := 47.6
	longitude := -120.9

	fmt.Printf("Querying tree coverage at (%.4f, %.4f)...\n", latitude, longitude)

	coverage, err := helpers.TreeCoverage(client, latitude, longitude)
	if err != nil {
		log.Fatalf("Failed to get tree coverage: %v", err)
	}

	fmt.Printf("SUCCESS! Tree Canopy Coverage: %.2f%%\n", coverage)

	// Test with fluent API
	fmt.Println("\n=== Testing Fluent API (Direct) ===")

	result, err := client.Image("USGS/NLCD_RELEASES/2023_REL/TCC/v2023-5").
		Select("NLCD_Percent_Tree_Canopy_Cover").
		ReduceRegion(
			ee.NewPoint(longitude, latitude),
			ee.ReducerFirst(),
			ee.Scale(30),
		).
		Compute(ctx)

	if err != nil {
		log.Fatalf("Failed to query with fluent API: %v", err)
	}

	fmt.Printf("SUCCESS! Result: %v\n", result)

	// Test other helpers
	fmt.Println("\n=== Testing Other Helpers ===")

	elevation, err := helpers.Elevation(client, latitude, longitude)
	if err != nil {
		log.Printf("Elevation query failed: %v", err)
	} else {
		fmt.Printf("Elevation: %.0f meters\n", elevation)
	}

	urban, err := helpers.IsUrban(client, latitude, longitude)
	if err != nil {
		log.Printf("Urban check failed: %v", err)
	} else {
		fmt.Printf("Is urban: %v\n", urban)
	}
}
