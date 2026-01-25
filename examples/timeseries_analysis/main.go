package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alexscott64/go-earthengine"
	"github.com/alexscott64/go-earthengine/helpers"
)

// This example demonstrates time-series analysis capabilities.
//
// Use cases:
// - Vegetation trend analysis over time
// - Detecting seasonal patterns
// - Anomaly detection in environmental data
// - Change detection between time periods
// - Climate trend analysis

func main() {
	ctx := context.Background()

	// Initialize Earth Engine client
	// Note: This example requires proper authentication setup
	client := &earthengine.Client{}
	_ = ctx // For demonstration only

	fmt.Println("Time Series Analysis Examples")
	fmt.Println("==============================")
	fmt.Println()

	// Example 1: Trend analysis
	example1_TrendAnalysis(ctx, client)

	// Example 2: Anomaly detection
	example2_AnomalyDetection(ctx, client)

	// Example 3: Seasonal decomposition
	example3_SeasonalDecomposition(ctx, client)

	// Example 4: Change detection
	example4_ChangeDetection(ctx, client)

	// Example 5: Time series aggregation
	example5_Aggregation(ctx, client)
}

func example1_TrendAnalysis(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 1: Vegetation Trend Analysis")
	fmt.Println("-------------------------------------")

	// Create sample NDVI time series
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	points := make([]helpers.TimeSeriesPoint, 0)

	// Simulate increasing NDVI trend over 3 years
	for i := 0; i < 36; i++ {
		t := baseTime.AddDate(0, i, 0)
		// Linear trend + seasonal variation + noise
		trend := 0.5 + float64(i)*0.01
		seasonal := 0.1 * float64(i%12) / 12.0
		value := trend + seasonal

		points = append(points, helpers.TimeSeriesPoint{
			Time:  t,
			Value: value,
			Index: i,
		})
	}

	ts := &helpers.TimeSeries{
		Name:   "NDVI",
		Points: points,
	}

	// Analyze trend
	trend, err := helpers.AnalyzeTrend(ts)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Location: Agricultural region\n")
	fmt.Printf("Period: 2020-2023\n")
	fmt.Printf("\nTrend Analysis Results:\n")
	fmt.Printf("  Direction: %s\n", trend.TrendDirection)
	fmt.Printf("  Slope: %.4f per day\n", trend.Slope)
	fmt.Printf("  R²: %.3f\n", trend.RSquared)
	fmt.Printf("  P-value: %.4f\n", trend.PValue)
	fmt.Printf("  Total Change: %.1f%%\n", trend.ChangePercent)
	fmt.Printf("  Start Value: %.3f\n", trend.StartValue)
	fmt.Printf("  End Value: %.3f\n", trend.EndValue)

	if trend.SignificantDiff {
		fmt.Printf("  ✓ Trend is statistically significant (p < 0.05)\n")
	} else {
		fmt.Printf("  ⚠ Trend is not statistically significant\n")
	}

	fmt.Println()

	// Interpretation
	if trend.TrendDirection == "increasing" {
		fmt.Println("Interpretation:")
		fmt.Println("  The vegetation index shows a positive trend, indicating")
		fmt.Println("  increasing vegetation health or biomass over the period.")
	}

	fmt.Println()
}

func example2_AnomalyDetection(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 2: Anomaly Detection")
	fmt.Println("-----------------------------")

	// Create time series with an anomaly
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	points := make([]helpers.TimeSeriesPoint, 0)

	for i := 0; i < 20; i++ {
		t := baseTime.AddDate(0, 0, i)
		value := 100.0

		// Insert anomaly at day 10
		if i == 10 {
			value = 200.0
		}

		points = append(points, helpers.TimeSeriesPoint{
			Time:  t,
			Value: value,
			Index: i,
		})
	}

	ts := &helpers.TimeSeries{
		Name:   "Temperature",
		Points: points,
	}

	// Detect anomalies using 3 standard deviations threshold
	anomalies := helpers.DetectAnomalies(ts, 3.0)

	fmt.Printf("Dataset: Daily temperature readings\n")
	fmt.Printf("Period: 20 days\n")
	fmt.Printf("Threshold: 3 standard deviations\n")
	fmt.Println()

	// Report anomalies
	anomalyCount := 0
	for _, a := range anomalies {
		if a.IsAnomaly {
			anomalyCount++
			fmt.Printf("⚠ Anomaly detected at %s:\n",
				a.Time.Format("2006-01-02"))
			fmt.Printf("  Value: %.1f\n", a.Value)
			fmt.Printf("  Z-score: %.2f\n", a.ZScore)
			fmt.Printf("  Deviation: %.1f standard deviations\n", a.Deviation)
			fmt.Println()
		}
	}

	if anomalyCount == 0 {
		fmt.Println("✓ No anomalies detected")
	} else {
		fmt.Printf("Found %d anomalies out of %d observations\n",
			anomalyCount, len(anomalies))
	}

	fmt.Println()
}

func example3_SeasonalDecomposition(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 3: Seasonal Decomposition")
	fmt.Println("----------------------------------")

	// Create time series with trend and seasonality
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	points := make([]helpers.TimeSeriesPoint, 0)

	for i := 0; i < 36; i++ { // 3 years of monthly data
		t := baseTime.AddDate(0, i, 0)
		// Trend + seasonal component
		trend := 100.0 + float64(i)*2.0
		seasonal := 20.0 * (float64(i%12) / 12.0)
		value := trend + seasonal

		points = append(points, helpers.TimeSeriesPoint{
			Time:  t,
			Value: value,
			Index: i,
		})
	}

	ts := &helpers.TimeSeries{
		Name:   "NDVI",
		Points: points,
	}

	// Decompose with 12-month period
	decomp, err := helpers.DecomposeTimeSeries(ts, 12)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Dataset: Monthly NDVI\n")
	fmt.Printf("Period: 3 years (36 months)\n")
	fmt.Printf("Seasonal Period: 12 months\n")
	fmt.Println()

	fmt.Println("Decomposition Components:")
	fmt.Printf("  Trend values: %d\n", len(decomp.Trend))
	fmt.Printf("  Seasonal values: %d\n", len(decomp.Seasonal))
	fmt.Printf("  Residual values: %d\n", len(decomp.Residual))
	fmt.Println()

	// Show sample values
	fmt.Println("Sample Values (first 6 months):")
	fmt.Println("Month | Original | Trend   | Seasonal | Residual")
	fmt.Println("------|----------|---------|----------|----------")
	for i := 0; i < 6 && i < len(points); i++ {
		fmt.Printf("%-5d | %8.2f | %7.2f | %8.2f | %8.2f\n",
			i+1,
			points[i].Value,
			decomp.Trend[i],
			decomp.Seasonal[i],
			decomp.Residual[i])
	}

	fmt.Println()
}

func example4_ChangeDetection(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 4: Change Detection")
	fmt.Println("---------------------------")

	// Create "before" time series
	beforePoints := make([]helpers.TimeSeriesPoint, 0)
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		t := baseTime.AddDate(0, 0, i)
		beforePoints = append(beforePoints, helpers.TimeSeriesPoint{
			Time:  t,
			Value: 0.5 + float64(i)*0.01,
			Index: i,
		})
	}

	before := &helpers.TimeSeries{
		Name:   "Before",
		Points: beforePoints,
	}

	// Create "after" time series (with significant increase)
	afterPoints := make([]helpers.TimeSeriesPoint, 0)
	afterTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		t := afterTime.AddDate(0, 0, i)
		afterPoints = append(afterPoints, helpers.TimeSeriesPoint{
			Time:  t,
			Value: 0.7 + float64(i)*0.01,
			Index: i,
		})
	}

	after := &helpers.TimeSeries{
		Name:   "After",
		Points: afterPoints,
	}

	// Detect change
	change, err := helpers.DetectChange(before, after)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Analysis: Forest Recovery After Fire\n")
	fmt.Printf("Before Period: 2020\n")
	fmt.Printf("After Period: 2021\n")
	fmt.Println()

	fmt.Println("Change Detection Results:")
	fmt.Printf("  Before Mean: %.3f\n", change.BeforeMean)
	fmt.Printf("  After Mean: %.3f\n", change.AfterMean)
	fmt.Printf("  Difference: %.3f\n", change.Difference)
	fmt.Printf("  Percent Change: %.1f%%\n", change.PercentDiff)
	fmt.Printf("  Direction: %s\n", change.Direction)
	fmt.Printf("  T-value: %.3f\n", change.TValue)
	fmt.Printf("  P-value: %.4f\n", change.PValue)

	if change.Significant {
		fmt.Printf("  ✓ Change is statistically significant\n")
	} else {
		fmt.Printf("  ⚠ Change is not statistically significant\n")
	}

	fmt.Println()

	// Interpretation
	fmt.Println("Interpretation:")
	if change.Direction == "increase" {
		fmt.Println("  Vegetation has recovered, showing increased NDVI values.")
		fmt.Println("  This indicates regrowth and improved vegetation health.")
	}

	fmt.Println()
}

func example5_Aggregation(ctx context.Context, client *earthengine.Client) {
	fmt.Println("Example 5: Time Series Aggregation")
	fmt.Println("-----------------------------------")

	// Create daily time series
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	points := make([]helpers.TimeSeriesPoint, 0)

	for i := 0; i < 90; i++ { // 90 days
		t := baseTime.AddDate(0, 0, i)
		value := 20.0 + float64(i%30) // Daily variation
		points = append(points, helpers.TimeSeriesPoint{
			Time:  t,
			Value: value,
			Index: i,
		})
	}

	dailyTS := &helpers.TimeSeries{
		Name:   "Daily Temperature",
		Points: points,
	}

	fmt.Printf("Original: %d daily observations\n", len(dailyTS.Points))
	fmt.Println()

	// Aggregate to monthly means
	monthlyMean, err := helpers.AggregateTimeSeries(dailyTS, "month",
		helpers.AggMean)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Monthly Mean Aggregation: %d months\n", len(monthlyMean.Points))
	for i, p := range monthlyMean.Points {
		fmt.Printf("  Month %d: %.2f\n", i+1, p.Value)
	}
	fmt.Println()

	// Aggregate to monthly max
	monthlyMax, err := helpers.AggregateTimeSeries(dailyTS, "month",
		helpers.AggMax)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Monthly Max Aggregation: %d months\n", len(monthlyMax.Points))
	for i, p := range monthlyMax.Points {
		fmt.Printf("  Month %d: %.2f\n", i+1, p.Value)
	}
	fmt.Println()

	// Filter to specific time range
	startDate := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC)
	filtered := helpers.FilterTimeRange(dailyTS, startDate, endDate)

	fmt.Printf("Filtered to February: %d days\n", len(filtered.Points))
	fmt.Printf("  First: %s (%.2f)\n",
		filtered.Points[0].Time.Format("2006-01-02"),
		filtered.Points[0].Value)
	fmt.Printf("  Last: %s (%.2f)\n",
		filtered.Points[len(filtered.Points)-1].Time.Format("2006-01-02"),
		filtered.Points[len(filtered.Points)-1].Value)

	fmt.Println()
}

// Additional helper functions for real-world scenarios

func analyzeVegetationTrend(ctx context.Context, client *earthengine.Client, lat, lon float64, startDate, endDate string) (*helpers.TrendResult, error) {
	// In a real implementation:
	// 1. Get NDVI time series from Earth Engine
	// 2. Extract values for the location
	// 3. Analyze trend

	// Placeholder
	_ = ctx
	_ = client
	_ = lat
	_ = lon
	_ = startDate
	_ = endDate

	return &helpers.TrendResult{
		TrendDirection:  "increasing",
		Slope:           0.001,
		RSquared:        0.85,
		ChangePercent:   15.5,
		SignificantDiff: true,
	}, nil
}

func detectDroughtEvents(ctx context.Context, client *earthengine.Client, region *earthengine.Geometry, startDate, endDate string) ([]helpers.AnomalyResult, error) {
	// In a real implementation:
	// 1. Get precipitation or soil moisture time series
	// 2. Detect anomalously low values
	// 3. Identify drought events

	// Placeholder
	_ = ctx
	_ = client
	_ = region
	_ = startDate
	_ = endDate

	return []helpers.AnomalyResult{}, nil
}

func compareSeasonsAcrossYears(ctx context.Context, client *earthengine.Client, lat, lon float64, years []int) (map[int]*helpers.SeasonalDecomposition, error) {
	// In a real implementation:
	// 1. Get time series for each year
	// 2. Decompose to extract seasonal patterns
	// 3. Compare seasonal components

	// Placeholder
	_ = ctx
	_ = client
	_ = lat
	_ = lon

	results := make(map[int]*helpers.SeasonalDecomposition)
	for _, year := range years {
		results[year] = &helpers.SeasonalDecomposition{
			Period: 12,
		}
	}

	return results, nil
}
