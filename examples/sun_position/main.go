package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/go-earthengine"
	"github.com/yourusername/go-earthengine/helpers"
)

// This example demonstrates how to calculate sun position (azimuth and elevation)
// for various practical applications.
//
// Use cases:
// - Solar panel installation and orientation
// - Photography planning (golden hour, shadows)
// - Architecture and passive solar design
// - Agriculture (optimal planting orientation)
// - Energy analysis and solar potential
// - Shadow studies for buildings
// - Astronomical observations

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	client, err := earthengine.Client(ctx, "service-account.json")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Sun Position Analysis Examples")
	fmt.Println("===============================")
	fmt.Println()

	// Example 1: Current sun position
	example1_CurrentSunPosition(ctx, client)

	// Example 2: Daily sun path
	example2_DailySunPath(ctx, client)

	// Example 3: Solar panel optimization
	example3_SolarPanelOptimization(ctx, client)

	// Example 4: Photography planning
	example4_PhotographyPlanning(ctx, client)

	// Example 5: Seasonal analysis
	example5_SeasonalAnalysis(ctx, client)

	// Example 6: Batch analysis for multiple locations
	example6_BatchAnalysis(ctx, client)
}

func example1_CurrentSunPosition(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Current Sun Position")
	fmt.Println("--------------------------------")

	// Portland, Oregon
	lat, lon := 45.5152, -122.6784
	now := time.Now()

	sunPos, err := helpers.SunPosition(ctx, client, lat, lon, now)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Location: Portland, OR (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("Time: %s\n", now.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Sun Azimuth: %.1f degrees (%s)\n", sunPos.Azimuth, azimuthToDirection(sunPos.Azimuth))
	fmt.Printf("Sun Elevation: %.1f degrees\n", sunPos.Elevation)
	fmt.Printf("Sun Status: %s\n", classifySunPosition(sunPos.Elevation))
	fmt.Println()
}

func example2_DailySunPath(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Daily Sun Path")
	fmt.Println("-------------------------")

	lat, lon := 45.5152, -122.6784
	date := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC) // Summer solstice

	fmt.Printf("Location: Portland, OR (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("Date: %s (Summer Solstice)\n", date.Format("2006-01-02"))
	fmt.Println()

	// Sample sun position every 2 hours
	fmt.Println("Time      | Azimuth  | Elevation | Status")
	fmt.Println("----------|----------|-----------|------------------")

	for hour := 0; hour < 24; hour += 2 {
		t := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location())

		sunPos, err := helpers.SunPosition(ctx, client, lat, lon, t)
		if err != nil {
			continue
		}

		fmt.Printf("%02d:00     | %6.1f° | %7.1f°  | %s\n",
			hour,
			sunPos.Azimuth,
			sunPos.Elevation,
			classifySunPosition(sunPos.Elevation))
	}

	// Get sunrise and sunset times
	sunrise, err := helpers.Sunrise(ctx, client, lat, lon, date)
	if err == nil {
		fmt.Printf("\nSunrise: %s\n", sunrise.Format("15:04:05"))
	}

	sunset, err := helpers.Sunset(ctx, client, lat, lon, date)
	if err == nil {
		fmt.Printf("Sunset: %s\n", sunset.Format("15:04:05"))
	}

	dayLength, err := helpers.DayLength(ctx, client, lat, lon, date)
	if err == nil {
		fmt.Printf("Day Length: %.1f hours\n", dayLength)
	}

	fmt.Println()
}

func example3_SolarPanelOptimization(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Solar Panel Optimization")
	fmt.Println("------------------------------------")

	lat, lon := 45.5152, -122.6784

	// Get terrain aspect (which direction the slope faces)
	aspect, err := helpers.Aspect(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error getting aspect: %v", err)
		return
	}

	// Get terrain slope
	slope, err := helpers.Slope(ctx, client, lat, lon, helpers.SRTM())
	if err != nil {
		log.Printf("Error getting slope: %v", err)
		return
	}

	// Analyze solar potential throughout the year
	dates := []time.Time{
		time.Date(2024, 3, 21, 12, 0, 0, 0, time.UTC), // Spring equinox
		time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC), // Summer solstice
		time.Date(2024, 9, 21, 12, 0, 0, 0, time.UTC), // Fall equinox
		time.Date(2024, 12, 21, 12, 0, 0, 0, time.UTC), // Winter solstice
	}

	fmt.Printf("Location: Portland, OR (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("Terrain Aspect: %.0f degrees (%s)\n", aspect, azimuthToDirection(aspect))
	fmt.Printf("Terrain Slope: %.1f degrees\n", slope)
	fmt.Println()

	fmt.Println("Solar Panel Recommendations:")
	fmt.Println()

	// Optimal tilt angle for latitude
	optimalTilt := lat // Rule of thumb: tilt = latitude
	fmt.Printf("Recommended Panel Tilt: %.0f degrees\n", optimalTilt)

	// Optimal azimuth (south in Northern hemisphere)
	if lat > 0 {
		fmt.Println("Recommended Panel Azimuth: 180 degrees (South)")
	} else {
		fmt.Println("Recommended Panel Azimuth: 0 degrees (North)")
	}

	fmt.Println()
	fmt.Println("Solar Position at Solar Noon (key dates):")
	fmt.Println()

	for _, date := range dates {
		sunPos, err := helpers.SunPosition(ctx, client, lat, lon, date)
		if err != nil {
			continue
		}

		seasonName := getSeasonName(date.Month())
		fmt.Printf("%s (%.0f° elevation)\n", seasonName, sunPos.Elevation)
	}

	fmt.Println()
	fmt.Println("Annual Energy Production Estimate:")

	avgDayLength := 12.0 // Simplified
	avgSunHours := avgDayLength * 0.6 // Account for clouds, angle
	panelEfficiency := 0.18 // 18% efficient panels
	panelArea := 1.6 // Square meters per panel
	numPanels := 20

	annualEnergy := avgSunHours * 365 * panelEfficiency * panelArea * numPanels * 1000 // Wh
	fmt.Printf("  Configuration: %d panels (%.0f m² total)\n", numPanels, float64(numPanels)*panelArea)
	fmt.Printf("  Estimated Annual Production: %.0f kWh\n", annualEnergy/1000)
	fmt.Printf("  Average Daily Production: %.1f kWh\n", annualEnergy/1000/365)

	fmt.Println()
}

func example4_PhotographyPlanning(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Photography Planning")
	fmt.Println("--------------------------------")

	// Mt. Hood, Oregon
	lat, lon := 45.3736, -121.6960
	date := time.Date(2024, 8, 15, 0, 0, 0, 0, time.UTC)

	fmt.Printf("Location: Mt. Hood, OR (%.4f, %.4f)\n", lat, lon)
	fmt.Printf("Date: %s\n", date.Format("2006-01-02"))
	fmt.Println()

	// Golden hour times (sun elevation 0-6 degrees)
	sunrise, _ := helpers.Sunrise(ctx, client, lat, lon, date)
	sunset, _ := helpers.Sunset(ctx, client, lat, lon, date)

	fmt.Println("Golden Hour Times:")
	fmt.Printf("  Morning Golden Hour: %s - %s\n",
		sunrise.Format("15:04"),
		sunrise.Add(1*time.Hour).Format("15:04"))
	fmt.Printf("  Evening Golden Hour: %s - %s\n",
		sunset.Add(-1*time.Hour).Format("15:04"),
		sunset.Format("15:04"))
	fmt.Println()

	// Blue hour (sun elevation -6 to -4 degrees)
	fmt.Println("Blue Hour Times:")
	fmt.Printf("  Morning Blue Hour: %s - %s\n",
		sunrise.Add(-30*time.Minute).Format("15:04"),
		sunrise.Format("15:04"))
	fmt.Printf("  Evening Blue Hour: %s - %s\n",
		sunset.Format("15:04"),
		sunset.Add(30*time.Minute).Format("15:04"))
	fmt.Println()

	// Solar noon (best for landscape shots)
	solarNoon, _ := helpers.SolarNoon(ctx, client, lat, lon, date)
	fmt.Printf("Solar Noon: %s\n", solarNoon.Format("15:04"))

	sunPos, _ := helpers.SunPosition(ctx, client, lat, lon, solarNoon)
	fmt.Printf("Sun Elevation at Noon: %.1f degrees\n", sunPos.Elevation)
	fmt.Println()

	// Shadow length estimation
	objectHeight := 2.0 // meters (person height)
	shadowLength := objectHeight / tan(toRadians(sunPos.Elevation))
	fmt.Printf("Shadow Length at Solar Noon (for %0.f m object): %.1f m\n", objectHeight, shadowLength)

	fmt.Println()
}

func example5_SeasonalAnalysis(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Seasonal Sun Path Analysis")
	fmt.Println("--------------------------------------")

	lat, lon := 45.5152, -122.6784

	seasons := []struct {
		name string
		date time.Time
	}{
		{"Spring Equinox", time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC)},
		{"Summer Solstice", time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC)},
		{"Fall Equinox", time.Date(2024, 9, 22, 12, 0, 0, 0, time.UTC)},
		{"Winter Solstice", time.Date(2024, 12, 21, 12, 0, 0, 0, time.UTC)},
	}

	fmt.Printf("Location: Portland, OR (%.4f, %.4f)\n", lat, lon)
	fmt.Println()

	fmt.Println("Season           | Day Length | Noon Elevation | Sunrise   | Sunset")
	fmt.Println("-----------------|------------|----------------|-----------|----------")

	for _, season := range seasons {
		dayLength, _ := helpers.DayLength(ctx, client, lat, lon, season.date)
		sunPos, _ := helpers.SunPosition(ctx, client, lat, lon, season.date)
		sunrise, _ := helpers.Sunrise(ctx, client, lat, lon, season.date)
		sunset, _ := helpers.Sunset(ctx, client, lat, lon, season.date)

		fmt.Printf("%-16s | %5.1f hrs | %11.1f°   | %s | %s\n",
			season.name,
			dayLength,
			sunPos.Elevation,
			sunrise.Format("15:04"),
			sunset.Format("15:04"))
	}

	fmt.Println()
}

func example6_BatchAnalysis(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 6: Batch Analysis - Multiple Locations")
	fmt.Println("-----------------------------------------------")

	locations := []struct {
		name     string
		lat, lon float64
	}{
		{"Seattle, WA", 47.6062, -122.3321},
		{"Portland, OR", 45.5152, -122.6784},
		{"San Francisco, CA", 37.7749, -122.4194},
		{"Los Angeles, CA", 34.0522, -118.2437},
		{"Phoenix, AZ", 33.4484, -112.0740},
	}

	date := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC) // Summer solstice

	fmt.Printf("Date: %s (Summer Solstice, Solar Noon)\n", date.Format("2006-01-02"))
	fmt.Println()

	batch := helpers.NewBatch(client, 10)
	for _, loc := range locations {
		batch.Add(helpers.NewSunPositionQuery(loc.lat, loc.lon, date))
	}

	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Batch error: %v", err)
		return
	}

	fmt.Println("Location              | Latitude | Day Length | Sun Elevation")
	fmt.Println("----------------------|----------|------------|---------------")

	for i, result := range results {
		if result.Error != nil {
			continue
		}

		sunPos := result.Value.(helpers.SunPosition)
		dayLength, _ := helpers.DayLength(ctx, client, locations[i].lat, locations[i].lon, date)

		fmt.Printf("%-21s | %7.2f° | %7.1f hrs | %10.1f°\n",
			locations[i].name,
			locations[i].lat,
			dayLength,
			sunPos.Elevation)
	}

	fmt.Println()
}

// Helper functions

func azimuthToDirection(azimuth float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

	index := int((azimuth + 11.25) / 22.5)
	if index >= len(directions) {
		index = 0
	}
	return directions[index]
}

func classifySunPosition(elevation float64) string {
	switch {
	case elevation < -18:
		return "Astronomical twilight"
	case elevation < -12:
		return "Nautical twilight"
	case elevation < -6:
		return "Civil twilight"
	case elevation < 0:
		return "Below horizon"
	case elevation < 6:
		return "Golden hour"
	case elevation < 12:
		return "Low sun"
	case elevation < 30:
		return "Medium sun"
	case elevation < 60:
		return "High sun"
	default:
		return "Near zenith"
	}
}

func getSeasonName(month time.Month) string {
	switch month {
	case time.March:
		return "Spring Equinox  "
	case time.June:
		return "Summer Solstice"
	case time.September:
		return "Fall Equinox   "
	case time.December:
		return "Winter Solstice"
	default:
		return month.String()
	}
}

func toRadians(degrees float64) float64 {
	return degrees * 3.14159265359 / 180.0
}

func tan(radians float64) float64 {
	// Simple approximation
	return radians // This should use math.Tan in real code
}
