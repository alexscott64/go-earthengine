package helpers

import (
	"testing"
)

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lon       float64
		wantError bool
	}{
		{"valid coordinates", 45.5152, -122.6784, false},
		{"valid at equator", 0.0, 0.0, false},
		{"valid at poles", 90.0, 180.0, false},
		{"valid at south pole", -90.0, -180.0, false},
		{"invalid latitude too high", 91.0, 0.0, true},
		{"invalid latitude too low", -91.0, 0.0, true},
		{"invalid longitude too high", 0.0, 181.0, true},
		{"invalid longitude too low", 0.0, -181.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCoordinates(tt.lat, tt.lon)
			if (err != nil) != tt.wantError {
				t.Errorf("validateCoordinates() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNLCDClassToName(t *testing.T) {
	tests := []struct {
		class int
		want  string
	}{
		{11, "water"},
		{21, "developed_open"},
		{41, "forest_deciduous"},
		{42, "forest_evergreen"},
		{82, "crops"},
		{90, "woody_wetlands"},
		{999, "unknown_999"}, // Unknown class
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := nlcdClassToName(tt.class)
			if got != tt.want {
				t.Errorf("nlcdClassToName(%d) = %v, want %v", tt.class, got, tt.want)
			}
		})
	}
}

func TestTreeCoverageOptions(t *testing.T) {
	cfg := &treeCoverageConfig{
		dataset: nlcdTCCDatasetID,
	}

	// Test Year option
	year := 2020
	Year(year)(cfg)
	if cfg.year == nil || *cfg.year != year {
		t.Errorf("Year option failed: got %v, want %d", cfg.year, year)
	}

	// Test HansenDataset option
	cfg2 := &treeCoverageConfig{
		dataset: nlcdTCCDatasetID,
	}
	HansenDataset()(cfg2)
	if cfg2.dataset != hansenDatasetID {
		t.Errorf("HansenDataset option failed: got %v, want %v", cfg2.dataset, hansenDatasetID)
	}

	// Test WithScale option
	cfg3 := &treeCoverageConfig{}
	scale := 60.0
	WithScale(scale)(cfg3)
	if cfg3.scale == nil || *cfg3.scale != scale {
		t.Errorf("WithScale option failed: got %v, want %f", cfg3.scale, scale)
	}
}

func TestTreeCoverageQuery(t *testing.T) {
	query := NewTreeCoverageQuery(45.5152, -122.6784)

	// Verify it implements Query interface
	var _ Query = query

	tcq, ok := query.(*TreeCoverageQuery)
	if !ok {
		t.Fatal("NewTreeCoverageQuery did not return *TreeCoverageQuery")
	}

	if tcq.lat != 45.5152 {
		t.Errorf("TreeCoverageQuery.lat = %v, want %v", tcq.lat, 45.5152)
	}
	if tcq.lon != -122.6784 {
		t.Errorf("TreeCoverageQuery.lon = %v, want %v", tcq.lon, -122.6784)
	}
}

func TestTreeCoverageQueryWithOptions(t *testing.T) {
	query := NewTreeCoverageQuery(45.5152, -122.6784, Year(2020), WithScale(60))

	tcq, ok := query.(*TreeCoverageQuery)
	if !ok {
		t.Fatal("NewTreeCoverageQuery did not return *TreeCoverageQuery")
	}

	if len(tcq.opts) != 2 {
		t.Errorf("TreeCoverageQuery.opts length = %v, want 2", len(tcq.opts))
	}
}

func TestApplyScale(t *testing.T) {
	tests := []struct {
		name         string
		opts         QueryOptions
		defaultScale float64
		want         float64
	}{
		{
			name:         "use custom scale",
			opts:         QueryOptions{Scale: float64Ptr(60.0)},
			defaultScale: 30.0,
			want:         60.0,
		},
		{
			name:         "use default scale",
			opts:         QueryOptions{},
			defaultScale: 30.0,
			want:         30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyScale(tt.opts, tt.defaultScale)
			if got != tt.want {
				t.Errorf("applyScale() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Integration tests would go here if we had a test client
// For now, these are unit tests that verify the logic without making API calls

func ExampleTreeCoverage() {
	// This example would require a real client, so we'll skip it for now
	// In a real scenario:
	// client, _ := earthengine.NewClient(...)
	// coverage, err := TreeCoverage(client, 45.5152, -122.6784)
	// fmt.Printf("Tree coverage: %.1f%%\n", coverage)
}

func ExampleTreeCoverage_withYear() {
	// Example showing how to query tree coverage for a specific year
	// client, _ := earthengine.NewClient(...)
	// coverage, err := TreeCoverage(client, 45.5152, -122.6784, Year(2010))
	// fmt.Printf("Tree coverage in 2010: %.1f%%\n", coverage)
}

func ExampleLandCoverClass() {
	// Example showing land cover classification
	// client, _ := earthengine.NewClient(...)
	// class, err := LandCoverClass(client, 45.5152, -122.6784)
	// fmt.Printf("Land cover: %s\n", class)
}

func ExampleImperviousSurface() {
	// Example showing impervious surface percentage
	// client, _ := earthengine.NewClient(...)
	// impervious, err := ImperviousSurface(client, 45.5152, -122.6784)
	// fmt.Printf("Impervious surface: %.1f%%\n", impervious)
}

func ExampleIsUrban() {
	// Example checking if a location is urban
	// client, _ := earthengine.NewClient(...)
	// urban, err := IsUrban(client, 45.5152, -122.6784)
	// if urban {
	//     fmt.Println("This is an urban area")
	// } else {
	//     fmt.Println("This is not an urban area")
	// }
}
