package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ee "github.com/alexscott64/go-earthengine"
)

func main() {
	ctx := context.Background()

	// Create client using environment variables
	client, err := ee.NewClient(ctx,
		ee.WithServiceAccountEnv(),
		ee.WithProject(os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")),
	)
	if err != nil {
		log.Fatalf("Failed to create Earth Engine client: %v", err)
	}

	// Example: Using the fluent API to query NLCD data
	fmt.Println("=== Basic Query Example ===")
	fmt.Println("Querying NLCD tree canopy data using the fluent API")

	// Location in Washington state
	latitude := 47.6
	longitude := -120.9

	// Build and execute the query
	result, err := client.Image("USGS/NLCD/NLCD2016").
		Select("percent_tree_cover").
		ReduceRegion(
			ee.NewPoint(longitude, latitude),
			ee.ReducerFirst(),
			ee.Scale(30), // NLCD resolution is 30 meters
		).
		Compute(ctx)

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Location: (%.4f, %.4f)\n", latitude, longitude)
	fmt.Printf("Result: %v\n\n", result)

	// Example 2: Using different reducers
	fmt.Println("=== Different Reducer Example ===")

	// You can use different reducers for different analysis needs
	reducers := []struct {
		name    string
		reducer ee.Reducer
	}{
		{"First", ee.ReducerFirst()},
		{"Mean", ee.ReducerMean()},
		{"Min", ee.ReducerMin()},
		{"Max", ee.ReducerMax()},
	}

	for _, r := range reducers {
		result, err := client.Image("USGS/NLCD/NLCD2016").
			Select("percent_tree_cover").
			ReduceRegion(
				ee.NewPoint(longitude, latitude),
				r.reducer,
				ee.Scale(30),
			).
			Compute(ctx)

		if err != nil {
			log.Printf("Error with %s reducer: %v", r.name, err)
			continue
		}

		fmt.Printf("%s reducer: %v\n", r.name, result)
	}

	// Example 3: Multiple bands
	fmt.Println("\n=== Multiple Bands Example ===")

	multiResult, err := client.Image("USGS/NLCD/NLCD2016").
		Select("percent_tree_cover", "impervious").
		ReduceRegion(
			ee.NewPoint(longitude, latitude),
			ee.ReducerFirst(),
			ee.Scale(30),
		).
		Compute(ctx)

	if err != nil {
		log.Fatalf("Multi-band query failed: %v", err)
	}

	fmt.Printf("Tree Canopy: %v\n", multiResult["percent_tree_cover"])
	fmt.Printf("Impervious Surface: %v\n", multiResult["impervious"])
}
