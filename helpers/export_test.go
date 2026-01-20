package helpers

import (
	"testing"
)

func TestExportDescriptionOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportDescription("My Export")(cfg)

	if cfg.Description != "My Export" {
		t.Errorf("Description = %s, want My Export", cfg.Description)
	}
}

func TestExportToGCSOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportToGCS("my-bucket", "exports/")(cfg)

	if cfg.Destination != ExportToCloudStorage {
		t.Errorf("Destination = %s, want %s", cfg.Destination, ExportToCloudStorage)
	}
	if cfg.Bucket != "my-bucket" {
		t.Errorf("Bucket = %s, want my-bucket", cfg.Bucket)
	}
	if cfg.Prefix != "exports/" {
		t.Errorf("Prefix = %s, want exports/", cfg.Prefix)
	}
}

func TestExportToGoogleDriveOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportToGoogleDrive("MyFolder")(cfg)

	if cfg.Destination != ExportToDrive {
		t.Errorf("Destination = %s, want %s", cfg.Destination, ExportToDrive)
	}
	if cfg.Folder != "MyFolder" {
		t.Errorf("Folder = %s, want MyFolder", cfg.Folder)
	}
}

func TestExportToEEAssetOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportToEEAsset("users/myuser/myasset")(cfg)

	if cfg.Destination != ExportToAsset {
		t.Errorf("Destination = %s, want %s", cfg.Destination, ExportToAsset)
	}
	if cfg.AssetID != "users/myuser/myasset" {
		t.Errorf("AssetID = %s, want users/myuser/myasset", cfg.AssetID)
	}
}

func TestExportScaleOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportScale(10.0)(cfg)

	if cfg.Scale != 10.0 {
		t.Errorf("Scale = %f, want 10.0", cfg.Scale)
	}
}

func TestExportCRSOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportCRS("EPSG:3857")(cfg)

	if cfg.CRS != "EPSG:3857" {
		t.Errorf("CRS = %s, want EPSG:3857", cfg.CRS)
	}
}

func TestExportMaxPixelsOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportMaxPixels(1e10)(cfg)

	if cfg.MaxPixels != 1e10 {
		t.Errorf("MaxPixels = %d, want 1e10", cfg.MaxPixels)
	}
}

func TestExportFileFormatOption(t *testing.T) {
	cfg := &ExportConfig{}
	ExportFileFormat(GeoTIFF)(cfg)

	if cfg.Format != GeoTIFF {
		t.Errorf("Format = %s, want %s", cfg.Format, GeoTIFF)
	}
}

func TestValidateExportConfigMissingDescription(t *testing.T) {
	cfg := &ExportConfig{
		Destination: ExportToCloudStorage,
		Bucket:      "my-bucket",
	}

	err := validateExportConfig(cfg)
	if err == nil {
		t.Error("Expected error for missing description")
	}
}

func TestValidateExportConfigMissingBucket(t *testing.T) {
	cfg := &ExportConfig{
		Description: "Test",
		Destination: ExportToCloudStorage,
		Scale:       30,
		MaxPixels:   1e9,
	}

	err := validateExportConfig(cfg)
	if err == nil {
		t.Error("Expected error for missing bucket")
	}
}

func TestValidateExportConfigMissingAssetID(t *testing.T) {
	cfg := &ExportConfig{
		Description: "Test",
		Destination: ExportToAsset,
		Scale:       30,
		MaxPixels:   1e9,
	}

	err := validateExportConfig(cfg)
	if err == nil {
		t.Error("Expected error for missing asset ID")
	}
}

func TestValidateExportConfigInvalidScale(t *testing.T) {
	cfg := &ExportConfig{
		Description: "Test",
		Destination: ExportToDrive,
		Scale:       -10,
		MaxPixels:   1e9,
	}

	err := validateExportConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid scale")
	}
}

func TestValidateExportConfigInvalidMaxPixels(t *testing.T) {
	cfg := &ExportConfig{
		Description: "Test",
		Destination: ExportToDrive,
		Scale:       30,
		MaxPixels:   -1,
	}

	err := validateExportConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid maxPixels")
	}
}

func TestValidateExportConfigValidDrive(t *testing.T) {
	cfg := &ExportConfig{
		Description: "Test",
		Destination: ExportToDrive,
		Folder:      "MyFolder",
		Scale:       30,
		MaxPixels:   1e9,
	}

	err := validateExportConfig(cfg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestExportDestinationConstants(t *testing.T) {
	destinations := []ExportDestination{
		ExportToCloudStorage,
		ExportToDrive,
		ExportToAsset,
	}

	for _, dest := range destinations {
		if dest == "" {
			t.Error("Export destination constant is empty")
		}
	}
}

func TestExportFormatConstants(t *testing.T) {
	formats := []ExportFormat{
		GeoTIFF,
		TFRecord,
		CSV,
		KML,
		KMZ,
		SHP,
		MP4,
	}

	for _, format := range formats {
		if format == "" {
			t.Error("Export format constant is empty")
		}
	}
}
