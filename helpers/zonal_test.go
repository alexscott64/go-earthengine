package helpers

import (
	"context"
	"strings"
	"testing"

	"github.com/alexscott64/go-earthengine"
)

func TestCalculateZonalStats(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	config := ZonalStatsConfig{
		Statistics: []ZonalStatistic{Mean, StdDev},
		Scale:      30,
		Bands:      []string{"B4", "B8"},
		ZoneIDKey:  "id",
	}

	result, err := CalculateZonalStats(ctx, client, image, zones, config)
	if err != nil {
		t.Fatalf("CalculateZonalStats failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Scale != 30 {
		t.Errorf("Scale = %f, want 30", result.Scale)
	}

	if len(result.Statistics) != 2 {
		t.Errorf("Got %d statistics, want 2", len(result.Statistics))
	}
}

func TestCalculateZonalStatsDefaults(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	// Test with minimal config
	config := ZonalStatsConfig{}

	result, err := CalculateZonalStats(ctx, client, image, zones, config)
	if err != nil {
		t.Fatalf("CalculateZonalStats failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check defaults were applied
	if result.Scale != 30 {
		t.Errorf("Default scale = %f, want 30", result.Scale)
	}

	if len(result.Statistics) != 1 || result.Statistics[0] != Mean {
		t.Error("Default statistic should be Mean")
	}
}

func TestCalculateZonalStatsSingle(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}

	stats, err := CalculateZonalStatsSingle(ctx, client, image, nil,
		[]ZonalStatistic{Mean, Max}, 10)
	if err != nil {
		t.Fatalf("CalculateZonalStatsSingle failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats is nil")
	}
}

func TestZonalMean(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	result, err := ZonalMean(ctx, client, image, zones, 30)
	if err != nil {
		t.Fatalf("ZonalMean failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Statistics) != 1 || result.Statistics[0] != Mean {
		t.Error("Should only calculate Mean statistic")
	}
}

func TestZonalSum(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	result, err := ZonalSum(ctx, client, image, zones, 30)
	if err != nil {
		t.Fatalf("ZonalSum failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Statistics) != 1 || result.Statistics[0] != Sum {
		t.Error("Should only calculate Sum statistic")
	}
}

func TestCalculateZonalHistogram(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	histograms, err := CalculateZonalHistogram(ctx, client, image, zones,
		"B4", 256, 0, 10000, 30)
	if err != nil {
		t.Fatalf("ZonalHistogram failed: %v", err)
	}

	if histograms == nil {
		t.Fatal("Histograms is nil")
	}
}

func TestZonalFrequencyTable(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	freq, err := ZonalFrequencyTable(ctx, client, image, zones,
		"classification", 30)
	if err != nil {
		t.Fatalf("ZonalFrequencyTable failed: %v", err)
	}

	if freq == nil {
		t.Fatal("Frequency table is nil")
	}
}

func TestCalculateZonalPercentiles(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	percentiles := []float64{25, 50, 75}
	result, err := CalculateZonalPercentiles(ctx, client, image, zones, percentiles, 30)
	if err != nil {
		t.Fatalf("ZonalPercentiles failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestCalculateZonalTimeSeries(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}
	zones := &earthengine.FeatureCollection{}

	series, err := CalculateZonalTimeSeries(ctx, client, collection, zones,
		Mean, "NDVI", 30)
	if err != nil {
		t.Fatalf("ZonalTimeSeries failed: %v", err)
	}

	if series == nil {
		t.Fatal("Series is nil")
	}
}

func TestCalculateZonalComparison(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	imageBefore := &earthengine.Image{}
	imageAfter := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	comparison, err := CalculateZonalComparison(ctx, client, imageBefore, imageAfter,
		zones, Mean, 30)
	if err != nil {
		t.Fatalf("ZonalComparison failed: %v", err)
	}

	if comparison == nil {
		t.Fatal("Comparison is nil")
	}
}

func TestZonalCrossTabulation(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image1 := &earthengine.Image{}
	image2 := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	crosstab, err := ZonalCrossTabulation(ctx, client, image1, image2,
		zones, 30)
	if err != nil {
		t.Fatalf("ZonalCrossTabulation failed: %v", err)
	}

	if crosstab == nil {
		t.Fatal("Cross-tabulation is nil")
	}
}

func TestCalculateZonalCorrelation(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	correlation, err := CalculateZonalCorrelation(ctx, client, image, zones,
		"B4", "B8", 30)
	if err != nil {
		t.Fatalf("ZonalCorrelation failed: %v", err)
	}

	if correlation == nil {
		t.Fatal("Correlation is nil")
	}
}

func TestExportZonalStatsToCSV(t *testing.T) {
	result := &ZonalStatsResult{
		Zones: []ZonalStats{
			{
				ZoneID: 1,
				Stats: map[string]float64{
					"B4_mean": 0.5,
					"B8_mean": 0.7,
				},
			},
			{
				ZoneID: 2,
				Stats: map[string]float64{
					"B4_mean": 0.6,
					"B8_mean": 0.8,
				},
			},
		},
	}

	csv, err := ExportZonalStatsToCSV(result)
	if err != nil {
		t.Fatalf("ExportZonalStatsToCSV failed: %v", err)
	}

	if csv == "" {
		t.Fatal("CSV is empty")
	}

	// Check header
	if !strings.Contains(csv, "zone_id") {
		t.Error("CSV missing zone_id header")
	}

	// Check data rows
	lines := strings.Split(csv, "\n")
	if len(lines) < 3 { // Header + at least 2 data rows
		t.Errorf("CSV has only %d lines, expected at least 3", len(lines))
	}
}

func TestExportZonalStatsToCSVEmpty(t *testing.T) {
	// Test with nil result
	_, err := ExportZonalStatsToCSV(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}

	// Test with empty zones
	emptyResult := &ZonalStatsResult{
		Zones: []ZonalStats{},
	}
	_, err = ExportZonalStatsToCSV(emptyResult)
	if err == nil {
		t.Error("Expected error for empty zones")
	}
}

func TestZonalStatsToFeatureCollection(t *testing.T) {
	result := &ZonalStatsResult{
		Zones: []ZonalStats{
			{ZoneID: 1, Stats: map[string]float64{"mean": 0.5}},
		},
	}

	fc, err := ZonalStatsToFeatureCollection(result)
	if err != nil {
		t.Fatalf("ZonalStatsToFeatureCollection failed: %v", err)
	}

	if fc == nil {
		t.Fatal("FeatureCollection is nil")
	}
}

func TestZonalStatsToFeatureCollectionNil(t *testing.T) {
	_, err := ZonalStatsToFeatureCollection(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestCalculateZonalStatsBatch(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	zones := &earthengine.FeatureCollection{}

	images := []*earthengine.Image{
		{},
		{},
		{},
	}

	config := ZonalStatsConfig{
		Statistics: []ZonalStatistic{Mean},
		Scale:      30,
	}

	results, err := CalculateZonalStatsBatch(ctx, client, images, zones, config)
	if err != nil {
		t.Fatalf("CalculateZonalStatsBatch failed: %v", err)
	}

	if len(results) != len(images) {
		t.Errorf("Got %d results, want %d", len(results), len(images))
	}

	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
		}
	}
}

func TestSummarizeZonalStats(t *testing.T) {
	result := &ZonalStatsResult{
		Zones: []ZonalStats{
			{
				ZoneID:     1,
				Stats:      map[string]float64{"B4_mean": 0.5},
				PixelCount: 100,
				Area:       1000.0,
			},
			{
				ZoneID:     2,
				Stats:      map[string]float64{"B4_mean": 0.7},
				PixelCount: 150,
				Area:       1500.0,
			},
			{
				ZoneID:     3,
				Stats:      map[string]float64{"B4_mean": 0.6},
				PixelCount: 120,
				Area:       1200.0,
			},
		},
	}

	summary := SummarizeZonalStats(result, "B4_mean")

	if summary == nil {
		t.Fatal("Summary is nil")
	}

	if summary.TotalZones != 3 {
		t.Errorf("TotalZones = %d, want 3", summary.TotalZones)
	}

	expectedMean := 0.6 // (0.5 + 0.7 + 0.6) / 3
	if summary.MeanValue < expectedMean-0.01 || summary.MeanValue > expectedMean+0.01 {
		t.Errorf("MeanValue = %f, want ~%f", summary.MeanValue, expectedMean)
	}

	if summary.MinValue != 0.5 {
		t.Errorf("MinValue = %f, want 0.5", summary.MinValue)
	}

	if summary.MaxValue != 0.7 {
		t.Errorf("MaxValue = %f, want 0.7", summary.MaxValue)
	}

	expectedArea := 3700.0 // 1000 + 1500 + 1200
	if summary.TotalArea != expectedArea {
		t.Errorf("TotalArea = %f, want %f", summary.TotalArea, expectedArea)
	}

	expectedPixels := 370 // 100 + 150 + 120
	if summary.TotalPixelCount != expectedPixels {
		t.Errorf("TotalPixelCount = %d, want %d", summary.TotalPixelCount, expectedPixels)
	}
}

func TestSummarizeZonalStatsEmpty(t *testing.T) {
	// Test with nil result
	summary := SummarizeZonalStats(nil, "B4_mean")
	if summary == nil {
		t.Fatal("Summary should not be nil")
	}
	if summary.TotalZones != 0 {
		t.Error("Empty result should have 0 zones")
	}

	// Test with empty zones
	emptyResult := &ZonalStatsResult{Zones: []ZonalStats{}}
	summary = SummarizeZonalStats(emptyResult, "B4_mean")
	if summary.TotalZones != 0 {
		t.Error("Empty zones should have 0 zones")
	}
}

func TestSummarizeZonalStatsMissingStatistic(t *testing.T) {
	result := &ZonalStatsResult{
		Zones: []ZonalStats{
			{ZoneID: 1, Stats: map[string]float64{"B4_mean": 0.5}},
		},
	}

	// Request non-existent statistic
	summary := SummarizeZonalStats(result, "B8_mean")

	if summary == nil {
		t.Fatal("Summary should not be nil")
	}

	// Should return empty summary when statistic doesn't exist
	if summary.TotalZones != 0 {
		t.Error("Should have 0 zones for non-existent statistic")
	}
}

func TestZonalStatisticConstants(t *testing.T) {
	stats := []ZonalStatistic{
		Mean,
		Sum,
		Count,
		Min,
		Max,
		Median,
		StdDev,
		Variance,
	}

	for _, stat := range stats {
		if string(stat) == "" {
			t.Errorf("Statistic %v has empty string value", stat)
		}
	}
}

func TestZonalStatsConfigValidation(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	image := &earthengine.Image{}
	zones := &earthengine.FeatureCollection{}

	// Test that invalid config still works with defaults
	config := ZonalStatsConfig{
		Scale:      0,  // Should default to 30
		MaxPixels:  0,  // Should default to 1e8
		TileScale:  0,  // Should default to 1
		Statistics: []ZonalStatistic{}, // Should default to [Mean]
	}

	result, err := CalculateZonalStats(ctx, client, image, zones, config)
	if err != nil {
		t.Fatalf("CalculateZonalStats failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestZonalStatsStructs(t *testing.T) {
	// Test ZonalStats structure
	stats := ZonalStats{
		ZoneID:     123,
		Stats:      map[string]float64{"mean": 0.5},
		PixelCount: 100,
		Area:       1000.0,
		Geometry:   nil,
	}

	if stats.ZoneID != 123 {
		t.Error("ZoneID not set correctly")
	}

	if stats.Stats["mean"] != 0.5 {
		t.Error("Stats not set correctly")
	}

	// Test ZonalHistogram structure
	histogram := ZonalHistogram{
		ZoneID:     456,
		BandName:   "B4",
		Histogram:  map[float64]int{0.5: 100},
		BucketSize: 0.1,
		MinValue:   0.0,
		MaxValue:   1.0,
	}

	if histogram.ZoneID != 456 {
		t.Error("Histogram ZoneID not set correctly")
	}

	// Test ZonalFrequency structure
	freq := ZonalFrequency{
		ZoneID:      789,
		ClassCounts: map[int]int{1: 100, 2: 200},
		Dominant:    2,
		Diversity:   0.5,
	}

	if freq.ZoneID != 789 {
		t.Error("Frequency ZoneID not set correctly")
	}

	// Test ZonalPercentiles structure
	percentiles := ZonalPercentiles{
		ZoneID:      101,
		BandName:    "NDVI",
		Percentiles: map[float64]float64{50: 0.5, 90: 0.9},
	}

	if percentiles.ZoneID != 101 {
		t.Error("Percentiles ZoneID not set correctly")
	}

	// Test ZonalTimeSeries structure
	ts := ZonalTimeSeries{
		ZoneID:   202,
		BandName: "NDVI",
		Series: []ZonalTimeSeriesPoint{
			{Time: "2023-01-01", Value: 0.5},
		},
	}

	if ts.ZoneID != 202 {
		t.Error("TimeSeries ZoneID not set correctly")
	}

	// Test ZonalComparison structure
	comparison := ZonalComparison{
		ZoneID:         303,
		BandName:       "NDVI",
		BeforeValue:    0.5,
		AfterValue:     0.7,
		Difference:     0.2,
		PercentChange:  40.0,
		ChangeCategory: "increase",
	}

	if comparison.ZoneID != 303 {
		t.Error("Comparison ZoneID not set correctly")
	}

	// Test ZonalCrossTab structure
	crosstab := ZonalCrossTab{
		ZoneID:    404,
		Matrix:    map[string]map[string]int{"forest": {"urban": 10}},
		FromTotal: map[string]int{"forest": 100},
		ToTotal:   map[string]int{"urban": 50},
	}

	if crosstab.ZoneID != 404 {
		t.Error("CrossTab ZoneID not set correctly")
	}

	// Test ZonalCorrelation structure
	correlation := ZonalCorrelation{
		ZoneID:      505,
		Band1:       "B4",
		Band2:       "B8",
		Correlation: 0.85,
		RSquared:    0.72,
		Slope:       1.2,
		Intercept:   0.1,
	}

	if correlation.ZoneID != 505 {
		t.Error("Correlation ZoneID not set correctly")
	}
}
