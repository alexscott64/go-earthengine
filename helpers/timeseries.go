package helpers

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/alexscott64/go-earthengine"
)

// TimeSeriesPoint represents a single data point in a time series.
type TimeSeriesPoint struct {
	Time  time.Time
	Value float64
	Index int // Original index in collection
}

// TimeSeries represents a collection of time-series data points.
type TimeSeries struct {
	Points []TimeSeriesPoint
	Name   string
}

// TrendResult contains trend analysis results.
type TrendResult struct {
	Slope           float64 // Rate of change per day
	Intercept       float64
	RSquared        float64 // Coefficient of determination (0-1)
	PValue          float64 // Statistical significance
	TrendDirection  string  // "increasing", "decreasing", "stable"
	ChangePercent   float64 // Total percent change over period
	StartValue      float64
	EndValue        float64
	SignificantDiff bool // Whether change is statistically significant
}

// AnomalyResult contains anomaly detection results.
type AnomalyResult struct {
	Index      int
	Time       time.Time
	Value      float64
	ZScore     float64
	IsAnomaly  bool
	Threshold  float64 // Z-score threshold used
	Deviation  float64 // Standard deviations from mean
}

// SeasonalDecomposition contains seasonal decomposition components.
type SeasonalDecomposition struct {
	Trend      []float64
	Seasonal   []float64
	Residual   []float64
	Period     int // Seasonal period
}

// ChangeDetectionResult contains change detection results.
type ChangeDetectionResult struct {
	BeforeMean  float64
	AfterMean   float64
	Difference  float64
	PercentDiff float64
	TValue      float64
	PValue      float64
	Significant bool
	Direction   string // "increase", "decrease", "no change"
}

// AnalyzeTrend performs linear regression trend analysis on time-series data.
//
// Example:
//
//	trend, err := helpers.AnalyzeTrend(timeSeries)
//	fmt.Printf("Trend: %s (%.2f per day)\n", trend.TrendDirection, trend.Slope)
//	fmt.Printf("R²: %.3f, Change: %.1f%%\n", trend.RSquared, trend.ChangePercent)
func AnalyzeTrend(ts *TimeSeries) (*TrendResult, error) {
	if len(ts.Points) < 2 {
		return nil, fmt.Errorf("need at least 2 points for trend analysis")
	}

	// Sort by time
	points := make([]TimeSeriesPoint, len(ts.Points))
	copy(points, ts.Points)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Time.Before(points[j].Time)
	})

	// Convert times to days since first point
	n := len(points)
	x := make([]float64, n)
	y := make([]float64, n)
	baseTime := points[0].Time

	for i, p := range points {
		x[i] = p.Time.Sub(baseTime).Hours() / 24.0 // Days
		y[i] = p.Value
	}

	// Calculate linear regression
	slope, intercept, rSquared := linearRegression(x, y)

	// Calculate p-value (simplified t-test)
	pValue := calculatePValue(x, y, slope, intercept)

	// Determine trend direction
	// Use relative threshold: 1% change per day relative to mean
	meanValue := calculateMean(y)
	relativeThreshold := math.Abs(meanValue) * 0.01
	if relativeThreshold < 0.01 {
		relativeThreshold = 0.01 // Minimum absolute threshold
	}

	direction := "stable"
	if math.Abs(slope) > relativeThreshold {
		if slope > 0 {
			direction = "increasing"
		} else {
			direction = "decreasing"
		}
	}

	// Calculate percent change
	startValue := points[0].Value
	endValue := points[n-1].Value
	changePercent := 0.0
	if startValue != 0 {
		changePercent = ((endValue - startValue) / startValue) * 100
	}

	// Significant if p-value < 0.05
	significant := pValue < 0.05

	return &TrendResult{
		Slope:           slope,
		Intercept:       intercept,
		RSquared:        rSquared,
		PValue:          pValue,
		TrendDirection:  direction,
		ChangePercent:   changePercent,
		StartValue:      startValue,
		EndValue:        endValue,
		SignificantDiff: significant,
	}, nil
}

// DetectAnomalies identifies anomalous values using z-score method.
//
// Example:
//
//	anomalies := helpers.DetectAnomalies(timeSeries, 3.0)
//	for _, a := range anomalies {
//	    if a.IsAnomaly {
//	        fmt.Printf("Anomaly at %s: %.2f (%.1f std devs)\n",
//	            a.Time.Format("2006-01-02"), a.Value, a.Deviation)
//	    }
//	}
func DetectAnomalies(ts *TimeSeries, threshold float64) []AnomalyResult {
	if len(ts.Points) < 3 {
		return nil
	}

	// Calculate mean and standard deviation
	values := make([]float64, len(ts.Points))
	for i, p := range ts.Points {
		values[i] = p.Value
	}

	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	results := make([]AnomalyResult, len(ts.Points))
	for i, p := range ts.Points {
		zScore := 0.0
		if stdDev > 0 {
			zScore = (p.Value - mean) / stdDev
		}

		results[i] = AnomalyResult{
			Index:     i,
			Time:      p.Time,
			Value:     p.Value,
			ZScore:    zScore,
			IsAnomaly: math.Abs(zScore) > threshold,
			Threshold: threshold,
			Deviation: math.Abs(zScore),
		}
	}

	return results
}

// DecomposeTimeSeries performs seasonal decomposition.
//
// Example:
//
//	decomp, err := helpers.DecomposeTimeSeries(timeSeries, 12) // 12-month seasonality
//	fmt.Printf("Trend: %v\n", decomp.Trend)
//	fmt.Printf("Seasonal: %v\n", decomp.Seasonal)
func DecomposeTimeSeries(ts *TimeSeries, period int) (*SeasonalDecomposition, error) {
	if len(ts.Points) < period*2 {
		return nil, fmt.Errorf("need at least %d points for seasonal decomposition", period*2)
	}

	n := len(ts.Points)
	values := make([]float64, n)
	for i, p := range ts.Points {
		values[i] = p.Value
	}

	// Calculate trend using moving average
	trend := movingAverage(values, period)

	// Calculate seasonal component
	seasonal := make([]float64, n)
	detrended := make([]float64, n)
	for i := 0; i < n; i++ {
		detrended[i] = values[i] - trend[i]
	}

	// Average seasonal pattern
	seasonalPattern := make([]float64, period)
	counts := make([]int, period)
	for i := 0; i < n; i++ {
		idx := i % period
		seasonalPattern[idx] += detrended[i]
		counts[idx]++
	}
	for i := 0; i < period; i++ {
		if counts[i] > 0 {
			seasonalPattern[i] /= float64(counts[i])
		}
	}

	// Apply seasonal pattern
	for i := 0; i < n; i++ {
		seasonal[i] = seasonalPattern[i%period]
	}

	// Calculate residual
	residual := make([]float64, n)
	for i := 0; i < n; i++ {
		residual[i] = values[i] - trend[i] - seasonal[i]
	}

	return &SeasonalDecomposition{
		Trend:    trend,
		Seasonal: seasonal,
		Residual: residual,
		Period:   period,
	}, nil
}

// DetectChange detects significant change between two time periods.
//
// Example:
//
//	beforeSeries := helpers.FilterTimeRange(series, start1, end1)
//	afterSeries := helpers.FilterTimeRange(series, start2, end2)
//	change, err := helpers.DetectChange(beforeSeries, afterSeries)
//	fmt.Printf("Change: %s (%.1f%%)\n", change.Direction, change.PercentDiff)
func DetectChange(before, after *TimeSeries) (*ChangeDetectionResult, error) {
	if len(before.Points) < 2 || len(after.Points) < 2 {
		return nil, fmt.Errorf("need at least 2 points in each period")
	}

	// Calculate means
	beforeValues := make([]float64, len(before.Points))
	for i, p := range before.Points {
		beforeValues[i] = p.Value
	}
	beforeMean := calculateMean(beforeValues)

	afterValues := make([]float64, len(after.Points))
	for i, p := range after.Points {
		afterValues[i] = p.Value
	}
	afterMean := calculateMean(afterValues)

	// Calculate difference
	diff := afterMean - beforeMean
	percentDiff := 0.0
	if beforeMean != 0 {
		percentDiff = (diff / beforeMean) * 100
	}

	// Perform t-test
	tValue, pValue := tTest(beforeValues, afterValues)

	// Determine direction
	direction := "no change"
	if math.Abs(percentDiff) > 1.0 { // 1% threshold
		if diff > 0 {
			direction = "increase"
		} else {
			direction = "decrease"
		}
	}

	return &ChangeDetectionResult{
		BeforeMean:  beforeMean,
		AfterMean:   afterMean,
		Difference:  diff,
		PercentDiff: percentDiff,
		TValue:      tValue,
		PValue:      pValue,
		Significant: pValue < 0.05,
		Direction:   direction,
	}, nil
}

// AggregateTimeSeries aggregates time series by a given period.
//
// Example:
//
//	// Aggregate to monthly means
//	monthly := helpers.AggregateTimeSeries(daily, "month", helpers.AggMean)
func AggregateTimeSeries(ts *TimeSeries, period string, aggFunc AggregationFunc) (*TimeSeries, error) {
	if len(ts.Points) == 0 {
		return &TimeSeries{Name: ts.Name}, nil
	}

	// Sort by time
	points := make([]TimeSeriesPoint, len(ts.Points))
	copy(points, ts.Points)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Time.Before(points[j].Time)
	})

	// Group by period
	groups := make(map[string][]float64)
	groupTimes := make(map[string]time.Time)

	for _, p := range points {
		key := getPeriodKey(p.Time, period)
		groups[key] = append(groups[key], p.Value)
		if _, exists := groupTimes[key]; !exists {
			groupTimes[key] = p.Time
		}
	}

	// Aggregate each group
	result := &TimeSeries{
		Name:   ts.Name + "_" + period,
		Points: make([]TimeSeriesPoint, 0, len(groups)),
	}

	for key, values := range groups {
		aggValue := aggFunc(values)
		result.Points = append(result.Points, TimeSeriesPoint{
			Time:  groupTimes[key],
			Value: aggValue,
		})
	}

	// Sort result
	sort.Slice(result.Points, func(i, j int) bool {
		return result.Points[i].Time.Before(result.Points[j].Time)
	})

	return result, nil
}

// AggregationFunc is a function that aggregates a slice of values.
type AggregationFunc func([]float64) float64

// Common aggregation functions
var (
	AggMean = func(values []float64) float64 {
		return calculateMean(values)
	}

	AggMedian = func(values []float64) float64 {
		return calculateMedian(values)
	}

	AggSum = func(values []float64) float64 {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	}

	AggMin = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		return min
	}

	AggMax = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		return max
	}
)

// FilterTimeRange filters a time series to a specific date range.
//
// Example:
//
//	filtered := helpers.FilterTimeRange(series,
//	    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
//	    time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC))
func FilterTimeRange(ts *TimeSeries, start, end time.Time) *TimeSeries {
	result := &TimeSeries{
		Name:   ts.Name,
		Points: make([]TimeSeriesPoint, 0),
	}

	for _, p := range ts.Points {
		if (p.Time.Equal(start) || p.Time.After(start)) &&
			(p.Time.Equal(end) || p.Time.Before(end)) {
			result.Points = append(result.Points, p)
		}
	}

	return result
}

// TimeSeriesFromImageCollection extracts a time series from an ImageCollection.
//
// Example:
//
//	ts, err := helpers.TimeSeriesFromImageCollection(ctx, client, collection,
//	    lat, lon, "NDVI")
func TimeSeriesFromImageCollection(ctx context.Context, client *earthengine.Client, collection *earthengine.ImageCollection, lat, lon float64, bandName string) (*TimeSeries, error) {
	// In a real implementation, this would:
	// 1. Get the list of images in the collection
	// 2. Extract the value at the point for each image
	// 3. Get the timestamp for each image
	// 4. Build the time series

	// For now, return a placeholder
	_ = ctx
	_ = client
	_ = collection
	_ = lat
	_ = lon
	_ = bandName

	return &TimeSeries{
		Name:   bandName,
		Points: []TimeSeriesPoint{},
	}, nil
}

// Helper functions

func linearRegression(x, y []float64) (slope, intercept, rSquared float64) {
	n := float64(len(x))
	if n == 0 {
		return 0, 0, 0
	}

	// Calculate means
	sumX, sumY := 0.0, 0.0
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate slope and intercept
	numerator := 0.0
	denominator := 0.0
	for i := 0; i < len(x); i++ {
		numerator += (x[i] - meanX) * (y[i] - meanY)
		denominator += (x[i] - meanX) * (x[i] - meanX)
	}

	if denominator == 0 {
		return 0, meanY, 0
	}

	slope = numerator / denominator
	intercept = meanY - slope*meanX

	// Calculate R²
	ssRes := 0.0
	ssTot := 0.0
	for i := 0; i < len(x); i++ {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	if ssTot == 0 {
		rSquared = 1.0
	} else {
		rSquared = 1.0 - (ssRes / ssTot)
	}

	return slope, intercept, rSquared
}

func calculatePValue(x, y []float64, slope, intercept float64) float64 {
	n := float64(len(x))
	if n < 3 {
		return 1.0
	}

	// Calculate standard error
	sumSquaredResiduals := 0.0
	for i := 0; i < len(x); i++ {
		predicted := slope*x[i] + intercept
		residual := y[i] - predicted
		sumSquaredResiduals += residual * residual
	}

	mse := sumSquaredResiduals / (n - 2)

	// Calculate standard error of slope
	sumSquaredX := 0.0
	meanX := 0.0
	for _, xi := range x {
		meanX += xi
	}
	meanX /= n

	for _, xi := range x {
		sumSquaredX += (xi - meanX) * (xi - meanX)
	}

	if sumSquaredX == 0 {
		return 1.0
	}

	se := math.Sqrt(mse / sumSquaredX)

	// Calculate t-statistic
	if se == 0 {
		return 0.0
	}
	tStat := math.Abs(slope / se)

	// Approximate p-value (two-tailed)
	// Using simplified approximation
	df := n - 2
	pValue := 2.0 * (1.0 - tDistributionCDF(tStat, df))

	return pValue
}

func tDistributionCDF(t, df float64) float64 {
	// Simplified approximation of t-distribution CDF
	// For production, use a proper statistical library
	if df > 30 {
		// Approximate with normal distribution for large df
		return 0.5 * (1.0 + math.Erf(t/math.Sqrt2))
	}

	// Very rough approximation
	x := df / (df + t*t)
	return 1.0 - 0.5*math.Pow(x, df/2.0)
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2.0
	}
	return sorted[n/2]
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values) - 1)
	return math.Sqrt(variance)
}

func tTest(sample1, sample2 []float64) (tValue, pValue float64) {
	n1 := float64(len(sample1))
	n2 := float64(len(sample2))

	if n1 < 2 || n2 < 2 {
		return 0, 1.0
	}

	mean1 := calculateMean(sample1)
	mean2 := calculateMean(sample2)

	stdDev1 := calculateStdDev(sample1, mean1)
	stdDev2 := calculateStdDev(sample2, mean2)

	// Pooled standard deviation
	pooledVar := ((n1-1)*stdDev1*stdDev1 + (n2-1)*stdDev2*stdDev2) / (n1 + n2 - 2)
	pooledStd := math.Sqrt(pooledVar)

	if pooledStd == 0 {
		return 0, 1.0
	}

	// t-statistic
	tValue = (mean1 - mean2) / (pooledStd * math.Sqrt(1/n1+1/n2))

	// Degrees of freedom
	df := n1 + n2 - 2

	// Approximate p-value
	pValue = 2.0 * (1.0 - tDistributionCDF(math.Abs(tValue), df))

	return tValue, pValue
}

func movingAverage(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - period/2
		end := i + period/2 + 1

		if start < 0 {
			start = 0
		}
		if end > n {
			end = n
		}

		sum := 0.0
		count := 0
		for j := start; j < end; j++ {
			sum += values[j]
			count++
		}

		if count > 0 {
			result[i] = sum / float64(count)
		}
	}

	return result
}

func getPeriodKey(t time.Time, period string) string {
	switch period {
	case "day":
		return t.Format("2006-01-02")
	case "week":
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "month":
		return t.Format("2006-01")
	case "year":
		return t.Format("2006")
	default:
		return t.Format("2006-01-02")
	}
}
