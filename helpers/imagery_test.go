package helpers

import (
	"testing"
)

func TestImageryOptions(t *testing.T) {
	tests := []struct {
		name   string
		option ImageryOption
		want   string
	}{
		{"Landsat8", Landsat8(), landsat8DatasetID},
		{"Landsat9", Landsat9(), landsat9DatasetID},
		{"Sentinel2", Sentinel2(), sentinel2DatasetID},
		{"MODIS", MODIS(), modisVIDatasetID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &imageryConfig{
				dataset: landsat8DatasetID, // Default
			}
			tt.option(cfg)
			if cfg.dataset != tt.want {
				t.Errorf("option() dataset = %v, want %v", cfg.dataset, tt.want)
			}
		})
	}
}

func TestCloudMaskOption(t *testing.T) {
	cfg := &imageryConfig{}
	cloudPercent := 20.0
	CloudMask(cloudPercent)(cfg)

	if cfg.cloudCover == nil || *cfg.cloudCover != cloudPercent {
		t.Errorf("CloudMask() cloudCover = %v, want %f", cfg.cloudCover, cloudPercent)
	}
}

func TestDateRangeOption(t *testing.T) {
	cfg := &imageryConfig{}
	DateRangeOption("2023-06-01", "2023-08-31")(cfg)

	if cfg.dateRange == nil {
		t.Fatal("DateRangeOption() dateRange is nil")
	}
	if cfg.dateRange.Start != "2023-06-01" {
		t.Errorf("dateRange.Start = %v, want 2023-06-01", cfg.dateRange.Start)
	}
	if cfg.dateRange.End != "2023-08-31" {
		t.Errorf("dateRange.End = %v, want 2023-08-31", cfg.dateRange.End)
	}
}

func TestImageryWithScale(t *testing.T) {
	cfg := &imageryConfig{}
	scale := 10.0
	ImageryWithScale(scale)(cfg)

	if cfg.scale == nil || *cfg.scale != scale {
		t.Errorf("ImageryWithScale() scale = %v, want %f", cfg.scale, scale)
	}
}

func TestNDVIQuery(t *testing.T) {
	query := NewNDVIQuery(45.5152, -122.6784, "2023-06-01")

	// Verify it implements Query interface
	var _ Query = query

	ndviQ, ok := query.(*NDVIQuery)
	if !ok {
		t.Fatal("NewNDVIQuery did not return *NDVIQuery")
	}

	if ndviQ.lat != 45.5152 {
		t.Errorf("NDVIQuery.lat = %v, want %v", ndviQ.lat, 45.5152)
	}
	if ndviQ.lon != -122.6784 {
		t.Errorf("NDVIQuery.lon = %v, want %v", ndviQ.lon, -122.6784)
	}
	if ndviQ.date != "2023-06-01" {
		t.Errorf("NDVIQuery.date = %v, want 2023-06-01", ndviQ.date)
	}
}

func TestNDVIQueryWithOptions(t *testing.T) {
	query := NewNDVIQuery(45.5152, -122.6784, "2023-06-01",
		Sentinel2(),
		CloudMask(20))

	ndviQ, ok := query.(*NDVIQuery)
	if !ok {
		t.Fatal("NewNDVIQuery did not return *NDVIQuery")
	}

	if len(ndviQ.opts) != 2 {
		t.Errorf("NDVIQuery.opts length = %v, want 2", len(ndviQ.opts))
	}
}

func TestEVIQuery(t *testing.T) {
	query := NewEVIQuery(45.5152, -122.6784, "2023-06-01")

	// Verify it implements Query interface
	var _ Query = query

	eviQ, ok := query.(*EVIQuery)
	if !ok {
		t.Fatal("NewEVIQuery did not return *EVIQuery")
	}

	if eviQ.lat != 45.5152 {
		t.Errorf("EVIQuery.lat = %v, want %v", eviQ.lat, 45.5152)
	}
	if eviQ.lon != -122.6784 {
		t.Errorf("EVIQuery.lon = %v, want %v", eviQ.lon, -122.6784)
	}
	if eviQ.date != "2023-06-01" {
		t.Errorf("EVIQuery.date = %v, want 2023-06-01", eviQ.date)
	}
}

func TestCompositeMethod(t *testing.T) {
	// Verify composite methods are defined
	methods := []CompositeMethod{
		MedianComposite,
		MeanComposite,
		MosaicComposite,
		GreenestPixelComposite,
	}

	for _, method := range methods {
		if method == "" {
			t.Errorf("CompositeMethod is empty")
		}
	}
}

func TestImageryDatasetConstants(t *testing.T) {
	// Verify dataset IDs are set correctly
	datasets := map[string]string{
		"Landsat8":  landsat8DatasetID,
		"Landsat9":  landsat9DatasetID,
		"Sentinel2": sentinel2DatasetID,
		"MODIS":     modisVIDatasetID,
	}

	for name, id := range datasets {
		if id == "" {
			t.Errorf("%s dataset ID is empty", name)
		}
	}
}

func TestSpectralBandsRequiresValidCoordinates(t *testing.T) {
	_, err := SpectralBands(nil, 100, -122, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNDVIRequiresValidCoordinates(t *testing.T) {
	_, err := NDVI(nil, -95, 200, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestEVIRequiresValidCoordinates(t *testing.T) {
	_, err := EVI(nil, 91, -122, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestSAVIRequiresValidCoordinates(t *testing.T) {
	_, err := SAVI(nil, 45.5, 200, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNDWIRequiresValidCoordinates(t *testing.T) {
	_, err := NDWI(nil, -95, -122, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNDBIRequiresValidCoordinates(t *testing.T) {
	_, err := NDBI(nil, 100, -122, "2023-06-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func ExampleNDVI() {
	// Example showing NDVI calculation
	// client, _ := earthengine.NewClient(...)
	// ndvi, err := NDVI(client, 45.5152, -122.6784, "2023-06-01",
	//     Sentinel2(),
	//     DateRangeOption("2023-06-01", "2023-08-31"),
	//     CloudMask(20))
	// fmt.Printf("NDVI: %.3f\n", ndvi)
}

func ExampleEVI() {
	// Example showing EVI calculation
	// client, _ := earthengine.NewClient(...)
	// evi, err := EVI(client, 45.5152, -122.6784, "2023-06-01",
	//     Landsat8())
	// fmt.Printf("EVI: %.3f\n", evi)
}

func ExampleSpectralBands() {
	// Example showing spectral band retrieval
	// client, _ := earthengine.NewClient(...)
	// bands, err := SpectralBands(client, 45.5152, -122.6784, "2023-06-01",
	//     Landsat8())
	// fmt.Printf("Red: %.4f, NIR: %.4f\n", bands["B4"], bands["B5"])
}

func ExampleComposite() {
	// Example showing composite creation
	// client, _ := earthengine.NewClient(...)
	// bounds := Bounds{MinLon: -123, MinLat: 45, MaxLon: -122, MaxLat: 46}
	// err := Composite(client, bounds,
	//     "2023-06-01", "2023-08-31",
	//     MedianComposite,
	//     Sentinel2(),
	//     CloudMask(20))
}
