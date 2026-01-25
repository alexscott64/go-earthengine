package helpers

import (
	"context"
	"testing"

	"github.com/alexscott64/go-earthengine"
)

func TestAdvancedComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	tests := []struct {
		name   string
		method CompositeMethod
	}{
		{"Median", MedianComposite},
		{"Mean", MeanComposite},
		{"Max", MaxComposite},
		{"Min", MinComposite},
		{"Percentile", PercentileComposite},
		{"Quality Mosaic", QualityMosaicComposite},
		{"Greenest Pixel", GreenestPixelComposite},
		{"Most Recent", MostRecentComposite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CompositeConfig{
				Method:         tt.method,
				CloudThreshold: 20,
				Percentile:     90,
			}

			result, err := AdvancedComposite(ctx, client, collection, config)
			if err != nil {
				t.Fatalf("AdvancedComposite failed: %v", err)
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Image == nil {
				t.Error("Result.Image is nil")
			}

			if result.Method != tt.method {
				t.Errorf("Method = %s, want %s", result.Method, tt.method)
			}
		})
	}
}

func TestAdvancedCompositeDefaults(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	// Test with empty config to verify defaults
	config := CompositeConfig{
		Method: MedianComposite,
	}

	result, err := AdvancedComposite(ctx, client, collection, config)
	if err != nil {
		t.Fatalf("AdvancedComposite failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestQualityMosaic(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	image, err := QualityMosaic(ctx, client, collection, "quality")
	if err != nil {
		t.Fatalf("QualityMosaic failed: %v", err)
	}

	if image == nil {
		t.Error("Image is nil")
	}
}

func TestGreenestPixelComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	image, err := CreateGreenestPixelComposite(ctx, client, collection)
	if err != nil {
		t.Fatalf("CreateGreenestPixelComposite failed: %v", err)
	}

	if image == nil {
		t.Error("Image is nil")
	}
}

func TestPercentileComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	// Test valid percentiles
	percentiles := []float64{10, 25, 50, 75, 90}
	for _, p := range percentiles {
		image, err := CreatePercentileComposite(ctx, client, collection, p)
		if err != nil {
			t.Errorf("CreatePercentileComposite(%f) failed: %v", p, err)
		}
		if image == nil {
			t.Errorf("Image is nil for percentile %f", p)
		}
	}

	// Test invalid percentiles
	invalidPercentiles := []float64{-10, 150}
	for _, p := range invalidPercentiles {
		_, err := CreatePercentileComposite(ctx, client, collection, p)
		if err == nil {
			t.Errorf("Expected error for invalid percentile %f", p)
		}
	}
}

func TestMostRecentComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	image, err := CreateMostRecentComposite(ctx, client, collection)
	if err != nil {
		t.Fatalf("CreateMostRecentComposite failed: %v", err)
	}

	if image == nil {
		t.Error("Image is nil")
	}
}

func TestSeasonalComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	seasons, err := SeasonalComposite(ctx, client, collection, 2023)
	if err != nil {
		t.Fatalf("SeasonalComposite failed: %v", err)
	}

	expectedSeasons := []string{"spring", "summer", "fall", "winter"}
	for _, season := range expectedSeasons {
		if _, exists := seasons[season]; !exists {
			t.Errorf("Missing season: %s", season)
		}

		if seasons[season] == nil {
			t.Errorf("Season %s image is nil", season)
		}
	}

	if len(seasons) != 4 {
		t.Errorf("Got %d seasons, want 4", len(seasons))
	}
}

func TestMultiTemporalComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	results, err := MultiTemporalComposite(ctx, client, collection,
		"2023-01-01", "2023-12-31", "month")
	if err != nil {
		t.Fatalf("MultiTemporalComposite failed: %v", err)
	}

	if results == nil {
		t.Fatal("Results is nil")
	}

	// Should have at least one result
	if len(results) == 0 {
		t.Error("No results returned")
	}

	// Check each result
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d is nil", i)
			continue
		}
		if result.Image == nil {
			t.Errorf("Result %d Image is nil", i)
		}
	}
}

func TestCompositeWithOutlierRemoval(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	thresholds := []float64{2.0, 2.5, 3.0, 3.5}
	for _, threshold := range thresholds {
		image, err := CompositeWithOutlierRemoval(ctx, client, collection, threshold)
		if err != nil {
			t.Errorf("CompositeWithOutlierRemoval(%f) failed: %v", threshold, err)
		}
		if image == nil {
			t.Errorf("Image is nil for threshold %f", threshold)
		}
	}
}

func TestPixelCompositeStats(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	result, err := PixelCompositeStats(ctx, client, collection)
	if err != nil {
		t.Fatalf("PixelCompositeStats failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Image == nil {
		t.Error("Result.Image is nil")
	}
}

func TestCompositeQualityMask(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	minObservations := []int{1, 3, 5, 10}
	for _, min := range minObservations {
		mask, err := CompositeQualityMask(ctx, client, collection, min)
		if err != nil {
			t.Errorf("CompositeQualityMask(%d) failed: %v", min, err)
		}
		if mask == nil {
			t.Errorf("Mask is nil for minObservations %d", min)
		}
	}
}

func TestCalculateCompositeMetrics(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	composite := &earthengine.Image{}

	metrics, err := CalculateCompositeMetrics(ctx, client, composite)
	if err != nil {
		t.Fatalf("CalculateCompositeMetrics failed: %v", err)
	}

	if metrics == nil {
		t.Fatal("Metrics is nil")
	}

	// Check that metrics have reasonable values
	if metrics.MeanObservations < 0 {
		t.Error("MeanObservations is negative")
	}

	if metrics.MedianObservations < 0 {
		t.Error("MedianObservations is negative")
	}

	if metrics.MinObservations < 0 {
		t.Error("MinObservations is negative")
	}

	if metrics.MaxObservations < metrics.MinObservations {
		t.Error("MaxObservations < MinObservations")
	}

	if metrics.Coverage < 0 || metrics.Coverage > 1 {
		t.Errorf("Coverage = %f, want 0-1", metrics.Coverage)
	}

	if metrics.CloudFreePixels < 0 || metrics.CloudFreePixels > 1 {
		t.Errorf("CloudFreePixels = %f, want 0-1", metrics.CloudFreePixels)
	}
}

func TestCompareComposites(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	composite1 := &earthengine.Image{}
	composite2 := &earthengine.Image{}

	diff, err := CompareComposites(ctx, client, composite1, composite2, "NDVI")
	if err != nil {
		t.Fatalf("CompareComposites failed: %v", err)
	}

	if diff == nil {
		t.Fatal("Diff is nil")
	}

	// Check that difference values are reasonable
	if diff.MaxDifference < diff.MeanDifference {
		t.Error("MaxDifference < MeanDifference")
	}

	if diff.PercentChanged < 0 || diff.PercentChanged > 100 {
		t.Errorf("PercentChanged = %f, want 0-100", diff.PercentChanged)
	}
}

func TestTemporalSmoothingComposite(t *testing.T) {
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	windowSizes := []int{3, 5, 7}
	for _, size := range windowSizes {
		images, err := TemporalSmoothingComposite(ctx, client, collection, size)
		if err != nil {
			t.Errorf("TemporalSmoothingComposite(%d) failed: %v", size, err)
		}
		if images == nil {
			t.Errorf("Images is nil for window size %d", size)
		}
	}
}

func TestCalculatePixelStats(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}

	stats := calculatePixelStats(values)

	if stats == nil {
		t.Fatal("Stats is nil")
	}

	// Check mean
	expectedMean := 30.0
	if stats["mean"] != expectedMean {
		t.Errorf("Mean = %f, want %f", stats["mean"], expectedMean)
	}

	// Check median
	expectedMedian := 30.0
	if stats["median"] != expectedMedian {
		t.Errorf("Median = %f, want %f", stats["median"], expectedMedian)
	}

	// Check min
	if stats["min"] != 10.0 {
		t.Errorf("Min = %f, want 10.0", stats["min"])
	}

	// Check max
	if stats["max"] != 50.0 {
		t.Errorf("Max = %f, want 50.0", stats["max"])
	}
}

func TestCalculatePixelStatsEmpty(t *testing.T) {
	stats := calculatePixelStats([]float64{})

	if stats == nil {
		t.Fatal("Stats is nil")
	}

	// All stats should be 0 for empty input
	for key, value := range stats {
		if value != 0 {
			t.Errorf("Stat %s = %f, want 0", key, value)
		}
	}
}

func TestCompositeConfigDefaults(t *testing.T) {
	config := CompositeConfig{
		Method: MedianComposite,
	}

	// Test that function applies defaults correctly
	ctx := context.Background()
	client := &earthengine.Client{}
	collection := &earthengine.ImageCollection{}

	result, err := AdvancedComposite(ctx, client, collection, config)
	if err != nil {
		t.Fatalf("AdvancedComposite failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestCompositeMethodString(t *testing.T) {
	methods := []CompositeMethod{
		MedianComposite,
		MeanComposite,
		MaxComposite,
		MinComposite,
		PercentileComposite,
		QualityMosaicComposite,
		GreenestPixelComposite,
		MostRecentComposite,
	}

	for _, method := range methods {
		if string(method) == "" {
			t.Errorf("Method %v has empty string value", method)
		}
	}
}

func TestDateRange(t *testing.T) {
	dr := DateRange{
		Start: "2023-01-01",
		End:   "2023-12-31",
	}

	if dr.Start == "" {
		t.Error("Start date is empty")
	}

	if dr.End == "" {
		t.Error("End date is empty")
	}
}

func TestGeneratePeriods(t *testing.T) {
	periods := generatePeriods("2023-01-01", "2023-12-31", "month")

	if len(periods) == 0 {
		t.Error("No periods generated")
	}

	// Check that periods have valid dates
	for i, p := range periods {
		if p.Start == "" {
			t.Errorf("Period %d has empty start date", i)
		}
		if p.End == "" {
			t.Errorf("Period %d has empty end date", i)
		}
	}
}
