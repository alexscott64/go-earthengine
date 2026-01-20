package helpers

import (
	"fmt"
	"math"
	"testing"
)

func TestElevationOptions(t *testing.T) {
	tests := []struct {
		name    string
		option  ElevationOption
		want    string
	}{
		{"SRTM", SRTM(), srtmDatasetID},
		{"ASTER", ASTER(), asterDatasetID},
		{"ALOS", ALOS(), alosDatasetID},
		{"USGS3DEP", USGS3DEP(), usgs3DEPDatasetID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &elevationConfig{
				dataset: srtmDatasetID, // Default
			}
			tt.option(cfg)
			if cfg.dataset != tt.want {
				t.Errorf("option() dataset = %v, want %v", cfg.dataset, tt.want)
			}
		})
	}
}

func TestElevationWithScale(t *testing.T) {
	cfg := &elevationConfig{}
	scale := 60.0
	ElevationWithScale(scale)(cfg)

	if cfg.scale == nil || *cfg.scale != scale {
		t.Errorf("ElevationWithScale() scale = %v, want %f", cfg.scale, scale)
	}
}

func TestElevationQuery(t *testing.T) {
	query := NewElevationQuery(39.7392, -104.9903)

	// Verify it implements Query interface
	var _ Query = query

	eq, ok := query.(*ElevationQuery)
	if !ok {
		t.Fatal("NewElevationQuery did not return *ElevationQuery")
	}

	if eq.lat != 39.7392 {
		t.Errorf("ElevationQuery.lat = %v, want %v", eq.lat, 39.7392)
	}
	if eq.lon != -104.9903 {
		t.Errorf("ElevationQuery.lon = %v, want %v", eq.lon, -104.9903)
	}
}

func TestElevationQueryWithOptions(t *testing.T) {
	query := NewElevationQuery(39.7392, -104.9903, USGS3DEP(), ElevationWithScale(10))

	eq, ok := query.(*ElevationQuery)
	if !ok {
		t.Fatal("NewElevationQuery did not return *ElevationQuery")
	}

	if len(eq.opts) != 2 {
		t.Errorf("ElevationQuery.opts length = %v, want 2", len(eq.opts))
	}
}

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		degrees float64
		want    float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{270, 3 * math.Pi / 2},
		{360, 2 * math.Pi},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.0f degrees", tt.degrees), func(t *testing.T) {
			got := degreesToRadians(tt.degrees)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("degreesToRadians(%v) = %v, want %v", tt.degrees, got, tt.want)
			}
		})
	}
}

func TestRadiansToDegrees(t *testing.T) {
	tests := []struct {
		radians float64
		want    float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{3 * math.Pi / 2, 270},
		{2 * math.Pi, 360},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f radians", tt.radians), func(t *testing.T) {
			got := radiansToDegrees(tt.radians)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("radiansToDegrees(%v) = %v, want %v", tt.radians, got, tt.want)
			}
		})
	}
}

func TestElevationDatasetConstants(t *testing.T) {
	// Verify dataset IDs are set correctly
	datasets := map[string]string{
		"SRTM":     srtmDatasetID,
		"ASTER":    asterDatasetID,
		"ALOS":     alosDatasetID,
		"USGS3DEP": usgs3DEPDatasetID,
	}

	for name, id := range datasets {
		if id == "" {
			t.Errorf("%s dataset ID is empty", name)
		}
	}
}

func TestElevationDefaultScales(t *testing.T) {
	// Verify default scales are reasonable
	scales := map[string]float64{
		"SRTM":     srtmDefaultScale,
		"ASTER":    asterDefaultScale,
		"ALOS":     alosDefaultScale,
		"USGS3DEP": usgs3DEPDefaultScale,
	}

	for name, scale := range scales {
		if scale <= 0 || scale > 1000 {
			t.Errorf("%s default scale %f is unreasonable", name, scale)
		}
	}
}

// Integration tests would go here if we had a test client

func ExampleElevation() {
	// Example showing basic elevation query
	// client, _ := earthengine.NewClient(...)
	// elev, err := Elevation(client, 39.7392, -104.9903)
	// fmt.Printf("Elevation: %.0f meters\n", elev)
}

func ExampleElevation_withDataset() {
	// Example showing elevation with different dataset
	// client, _ := earthengine.NewClient(...)
	// elev, err := Elevation(client, 39.7392, -104.9903, USGS3DEP())
	// fmt.Printf("Elevation (10m): %.0f meters\n", elev)
}

func ExampleSlope() {
	// Example showing slope calculation
	// client, _ := earthengine.NewClient(...)
	// slope, err := Slope(client, 39.7392, -104.9903)
	// fmt.Printf("Slope: %.1f degrees\n", slope)
}

func ExampleAspect() {
	// Example showing aspect calculation
	// client, _ := earthengine.NewClient(...)
	// aspect, err := Aspect(client, 39.7392, -104.9903)
	// fmt.Printf("Aspect: %.0f degrees\n", aspect)
}

func ExampleTerrainAnalysis() {
	// Example showing comprehensive terrain analysis
	// client, _ := earthengine.NewClient(...)
	// metrics, err := TerrainAnalysis(client, 39.7392, -104.9903)
	// fmt.Printf("Elevation: %.0fm\n", metrics.Elevation)
	// fmt.Printf("Slope: %.1f degrees\n", metrics.Slope)
	// fmt.Printf("Aspect: %.0f degrees\n", metrics.Aspect)
}
