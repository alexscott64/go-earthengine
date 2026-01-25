package helpers

import (
	"math"
	"testing"
	"time"
)

func TestAnalyzeTrend(t *testing.T) {
	// Create upward trending data
	ts := &TimeSeries{
		Name: "Test Trend",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 12.0},
			{Time: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 14.0},
			{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 16.0},
			{Time: time.Date(2023, 1, 5, 0, 0, 0, 0, time.UTC), Value: 18.0},
		},
	}

	trend, err := AnalyzeTrend(ts)
	if err != nil {
		t.Fatalf("AnalyzeTrend failed: %v", err)
	}

	if trend.TrendDirection != "increasing" {
		t.Errorf("TrendDirection = %s, want increasing", trend.TrendDirection)
	}

	if trend.Slope < 1.9 || trend.Slope > 2.1 {
		t.Errorf("Slope = %f, want ~2.0", trend.Slope)
	}

	if trend.RSquared < 0.99 {
		t.Errorf("RSquared = %f, want >= 0.99", trend.RSquared)
	}

	if trend.StartValue != 10.0 {
		t.Errorf("StartValue = %f, want 10.0", trend.StartValue)
	}

	if trend.EndValue != 18.0 {
		t.Errorf("EndValue = %f, want 18.0", trend.EndValue)
	}

	expectedChange := 80.0 // (18-10)/10 * 100
	if math.Abs(trend.ChangePercent-expectedChange) > 0.1 {
		t.Errorf("ChangePercent = %f, want %f", trend.ChangePercent, expectedChange)
	}
}

func TestAnalyzeTrendDecreasing(t *testing.T) {
	ts := &TimeSeries{
		Name: "Decreasing Trend",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 90.0},
			{Time: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 80.0},
			{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 70.0},
		},
	}

	trend, err := AnalyzeTrend(ts)
	if err != nil {
		t.Fatalf("AnalyzeTrend failed: %v", err)
	}

	if trend.TrendDirection != "decreasing" {
		t.Errorf("TrendDirection = %s, want decreasing", trend.TrendDirection)
	}

	if trend.Slope > -9 || trend.Slope < -11 {
		t.Errorf("Slope = %f, want ~-10.0", trend.Slope)
	}
}

func TestAnalyzeTrendStable(t *testing.T) {
	ts := &TimeSeries{
		Name: "Stable",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 50.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 50.1},
			{Time: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 49.9},
			{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 50.0},
		},
	}

	trend, err := AnalyzeTrend(ts)
	if err != nil {
		t.Fatalf("AnalyzeTrend failed: %v", err)
	}

	if trend.TrendDirection != "stable" {
		t.Errorf("TrendDirection = %s, want stable", trend.TrendDirection)
	}
}

func TestAnalyzeTrendInsufficientData(t *testing.T) {
	ts := &TimeSeries{
		Name:   "Too Short",
		Points: []TimeSeriesPoint{{Time: time.Now(), Value: 10.0}},
	}

	_, err := AnalyzeTrend(ts)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestDetectAnomalies(t *testing.T) {
	ts := &TimeSeries{
		Name: "With Anomalies",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 11.0},
			{Time: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 100.0}, // Anomaly
			{Time: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Value: 12.0},
			{Time: time.Date(2023, 1, 5, 0, 0, 0, 0, time.UTC), Value: 10.5},
		},
	}

	anomalies := DetectAnomalies(ts, 1.5) // Lower threshold to detect the anomaly

	if len(anomalies) != 5 {
		t.Errorf("Got %d results, want 5", len(anomalies))
	}

	// Check that index 2 is marked as anomaly
	if !anomalies[2].IsAnomaly {
		t.Error("Point at index 2 should be anomaly")
	}

	// Check that other points are not anomalies
	for i, a := range anomalies {
		if i == 2 {
			continue
		}
		if a.IsAnomaly {
			t.Errorf("Point at index %d should not be anomaly", i)
		}
	}

	// Check that anomaly has high z-score
	if anomalies[2].Deviation < 1.5 {
		t.Errorf("Anomaly deviation = %f, want >= 1.5", anomalies[2].Deviation)
	}
}

func TestDetectAnomaliesInsufficientData(t *testing.T) {
	ts := &TimeSeries{
		Name:   "Too Short",
		Points: []TimeSeriesPoint{{Time: time.Now(), Value: 10.0}},
	}

	anomalies := DetectAnomalies(ts, 3.0)
	if anomalies != nil {
		t.Error("Expected nil for insufficient data")
	}
}

func TestDecomposeTimeSeries(t *testing.T) {
	// Create data with clear seasonal pattern
	points := make([]TimeSeriesPoint, 24)
	for i := 0; i < 24; i++ {
		trend := float64(i) * 2.0                               // Linear trend
		seasonal := 10.0 * math.Sin(float64(i)*2.0*math.Pi/12) // 12-month cycle
		points[i] = TimeSeriesPoint{
			Time:  time.Date(2023, time.Month(i%12+1), 1, 0, 0, 0, 0, time.UTC),
			Value: trend + seasonal + 50.0,
		}
	}

	ts := &TimeSeries{
		Name:   "Seasonal",
		Points: points,
	}

	decomp, err := DecomposeTimeSeries(ts, 12)
	if err != nil {
		t.Fatalf("DecomposeTimeSeries failed: %v", err)
	}

	if len(decomp.Trend) != 24 {
		t.Errorf("Trend length = %d, want 24", len(decomp.Trend))
	}

	if len(decomp.Seasonal) != 24 {
		t.Errorf("Seasonal length = %d, want 24", len(decomp.Seasonal))
	}

	if len(decomp.Residual) != 24 {
		t.Errorf("Residual length = %d, want 24", len(decomp.Residual))
	}

	if decomp.Period != 12 {
		t.Errorf("Period = %d, want 12", decomp.Period)
	}
}

func TestDecomposeTimeSeriesInsufficientData(t *testing.T) {
	ts := &TimeSeries{
		Name: "Too Short",
		Points: []TimeSeriesPoint{
			{Time: time.Now(), Value: 10.0},
		},
	}

	_, err := DecomposeTimeSeries(ts, 12)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestDetectChange(t *testing.T) {
	before := &TimeSeries{
		Name: "Before",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 11.0},
			{Time: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Value: 10.5},
		},
	}

	after := &TimeSeries{
		Name: "After",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), Value: 20.0},
			{Time: time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC), Value: 21.0},
			{Time: time.Date(2023, 2, 3, 0, 0, 0, 0, time.UTC), Value: 19.5},
		},
	}

	change, err := DetectChange(before, after)
	if err != nil {
		t.Fatalf("DetectChange failed: %v", err)
	}

	if math.Abs(change.BeforeMean-10.5) > 0.1 {
		t.Errorf("BeforeMean = %f, want ~10.5", change.BeforeMean)
	}

	if math.Abs(change.AfterMean-20.167) > 0.1 {
		t.Errorf("AfterMean = %f, want ~20.167", change.AfterMean)
	}

	if change.Direction != "increase" {
		t.Errorf("Direction = %s, want increase", change.Direction)
	}

	if change.PercentDiff < 90 || change.PercentDiff > 95 {
		t.Errorf("PercentDiff = %f, want ~92", change.PercentDiff)
	}
}

func TestDetectChangeDecrease(t *testing.T) {
	before := &TimeSeries{
		Name: "Before",
		Points: []TimeSeriesPoint{
			{Time: time.Now(), Value: 100.0},
			{Time: time.Now(), Value: 110.0},
		},
	}

	after := &TimeSeries{
		Name: "After",
		Points: []TimeSeriesPoint{
			{Time: time.Now(), Value: 50.0},
			{Time: time.Now(), Value: 60.0},
		},
	}

	change, err := DetectChange(before, after)
	if err != nil {
		t.Fatalf("DetectChange failed: %v", err)
	}

	if change.Direction != "decrease" {
		t.Errorf("Direction = %s, want decrease", change.Direction)
	}
}

func TestDetectChangeInsufficientData(t *testing.T) {
	before := &TimeSeries{
		Name:   "Before",
		Points: []TimeSeriesPoint{{Time: time.Now(), Value: 10.0}},
	}

	after := &TimeSeries{
		Name:   "After",
		Points: []TimeSeriesPoint{{Time: time.Now(), Value: 20.0}},
	}

	_, err := DetectChange(before, after)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestAggregateTimeSeriesDay(t *testing.T) {
	ts := &TimeSeries{
		Name: "Hourly",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Value: 20.0},
			{Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Value: 15.0},
			{Time: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), Value: 25.0},
		},
	}

	daily, err := AggregateTimeSeries(ts, "day", AggMean)
	if err != nil {
		t.Fatalf("AggregateTimeSeries failed: %v", err)
	}

	if len(daily.Points) != 2 {
		t.Errorf("Got %d daily points, want 2", len(daily.Points))
	}

	// Check means
	if math.Abs(daily.Points[0].Value-15.0) > 0.1 {
		t.Errorf("Day 1 mean = %f, want 15.0", daily.Points[0].Value)
	}

	if math.Abs(daily.Points[1].Value-20.0) > 0.1 {
		t.Errorf("Day 2 mean = %f, want 20.0", daily.Points[1].Value)
	}
}

func TestAggregateTimeSeriesMonth(t *testing.T) {
	ts := &TimeSeries{
		Name: "Daily",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), Value: 20.0},
			{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), Value: 30.0},
			{Time: time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC), Value: 40.0},
		},
	}

	monthly, err := AggregateTimeSeries(ts, "month", AggMean)
	if err != nil {
		t.Fatalf("AggregateTimeSeries failed: %v", err)
	}

	if len(monthly.Points) != 2 {
		t.Errorf("Got %d monthly points, want 2", len(monthly.Points))
	}
}

func TestAggregationFunctions(t *testing.T) {
	values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}

	// Test mean
	mean := AggMean(values)
	if math.Abs(mean-30.0) > 0.01 {
		t.Errorf("Mean = %f, want 30.0", mean)
	}

	// Test median
	median := AggMedian(values)
	if math.Abs(median-30.0) > 0.01 {
		t.Errorf("Median = %f, want 30.0", median)
	}

	// Test sum
	sum := AggSum(values)
	if math.Abs(sum-150.0) > 0.01 {
		t.Errorf("Sum = %f, want 150.0", sum)
	}

	// Test min
	min := AggMin(values)
	if math.Abs(min-10.0) > 0.01 {
		t.Errorf("Min = %f, want 10.0", min)
	}

	// Test max
	max := AggMax(values)
	if math.Abs(max-50.0) > 0.01 {
		t.Errorf("Max = %f, want 50.0", max)
	}
}

func TestFilterTimeRange(t *testing.T) {
	ts := &TimeSeries{
		Name: "Full",
		Points: []TimeSeriesPoint{
			{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Value: 10.0},
			{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), Value: 20.0},
			{Time: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC), Value: 30.0},
			{Time: time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC), Value: 40.0},
		},
	}

	filtered := FilterTimeRange(ts,
		time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC))

	if len(filtered.Points) != 2 {
		t.Errorf("Got %d filtered points, want 2", len(filtered.Points))
	}

	if filtered.Points[0].Value != 20.0 {
		t.Errorf("First point value = %f, want 20.0", filtered.Points[0].Value)
	}

	if filtered.Points[1].Value != 30.0 {
		t.Errorf("Second point value = %f, want 30.0", filtered.Points[1].Value)
	}
}

func TestLinearRegression(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	slope, intercept, rSquared := linearRegression(x, y)

	if math.Abs(slope-2.0) > 0.01 {
		t.Errorf("Slope = %f, want 2.0", slope)
	}

	if math.Abs(intercept-0.0) > 0.01 {
		t.Errorf("Intercept = %f, want 0.0", intercept)
	}

	if math.Abs(rSquared-1.0) > 0.01 {
		t.Errorf("RSquared = %f, want 1.0", rSquared)
	}
}

func TestCalculateMean(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	mean := calculateMean(values)

	if math.Abs(mean-30.0) > 0.01 {
		t.Errorf("Mean = %f, want 30.0", mean)
	}

	// Empty slice
	emptyMean := calculateMean([]float64{})
	if emptyMean != 0 {
		t.Errorf("Empty mean = %f, want 0", emptyMean)
	}
}

func TestCalculateMedian(t *testing.T) {
	// Odd number of values
	values1 := []float64{10, 30, 20, 50, 40}
	median1 := calculateMedian(values1)
	if math.Abs(median1-30.0) > 0.01 {
		t.Errorf("Median (odd) = %f, want 30.0", median1)
	}

	// Even number of values
	values2 := []float64{10, 20, 30, 40}
	median2 := calculateMedian(values2)
	if math.Abs(median2-25.0) > 0.01 {
		t.Errorf("Median (even) = %f, want 25.0", median2)
	}

	// Empty slice
	emptyMedian := calculateMedian([]float64{})
	if emptyMedian != 0 {
		t.Errorf("Empty median = %f, want 0", emptyMedian)
	}
}

func TestCalculateStdDev(t *testing.T) {
	values := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	// Sample standard deviation with n-1: sqrt(32/7) ≈ 2.138
	expected := 2.138
	if math.Abs(stdDev-expected) > 0.01 {
		t.Errorf("StdDev = %f, want ~%f", stdDev, expected)
	}

	// Single value
	singleStdDev := calculateStdDev([]float64{10}, 10)
	if singleStdDev != 0 {
		t.Errorf("Single value stddev = %f, want 0", singleStdDev)
	}
}

func TestMovingAverage(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ma := movingAverage(values, 3)

	if len(ma) != len(values) {
		t.Errorf("Moving average length = %d, want %d", len(ma), len(values))
	}

	// Check middle value (should be average of 4, 5, 6)
	if math.Abs(ma[5]-6.0) > 0.5 {
		t.Errorf("MA[5] = %f, want ~6.0", ma[5])
	}
}

func TestGetPeriodKey(t *testing.T) {
	date := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

	dayKey := getPeriodKey(date, "day")
	if dayKey != "2023-06-15" {
		t.Errorf("Day key = %s, want 2023-06-15", dayKey)
	}

	monthKey := getPeriodKey(date, "month")
	if monthKey != "2023-06" {
		t.Errorf("Month key = %s, want 2023-06", monthKey)
	}

	yearKey := getPeriodKey(date, "year")
	if yearKey != "2023" {
		t.Errorf("Year key = %s, want 2023", yearKey)
	}
}
