package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// ExportDestination represents where to export data.
type ExportDestination string

const (
	// ExportToCloudStorage exports to Google Cloud Storage.
	ExportToCloudStorage ExportDestination = "CLOUD_STORAGE"
	// ExportToDrive exports to Google Drive.
	ExportToDrive ExportDestination = "DRIVE"
	// ExportToAsset exports to Earth Engine Asset.
	ExportToAsset ExportDestination = "ASSET"
)

// ExportFormat represents the export file format.
type ExportFormat string

const (
	// GeoTIFF format for raster exports.
	GeoTIFF ExportFormat = "GEO_TIFF"
	// TFRecord format for machine learning.
	TFRecord ExportFormat = "TF_RECORD"
	// CSV format for table exports.
	CSV ExportFormat = "CSV"
	// KML format for vector exports.
	KML ExportFormat = "KML"
	// KMZ format for compressed vector exports.
	KMZ ExportFormat = "KMZ"
	// SHP format for shapefiles.
	SHP ExportFormat = "SHP"
	// MP4 format for video exports.
	MP4 ExportFormat = "MP4"
)

// ExportConfig configures an export operation.
type ExportConfig struct {
	// Description of the export task
	Description string

	// Destination (Cloud Storage, Drive, or Asset)
	Destination ExportDestination

	// For Cloud Storage exports
	Bucket string
	Prefix string

	// For Drive exports
	Folder string

	// For Asset exports
	AssetID string

	// File format
	Format ExportFormat

	// Scale in meters per pixel
	Scale float64

	// CRS (coordinate reference system)
	CRS string

	// Region to export (geometry)
	Region *earthengine.Geometry

	// Max pixels to export
	MaxPixels int64

	// File dimensions
	FileDimensions []int
}

// ExportImageConfig creates an export configuration for image exports.
type ExportImageOption func(*ExportConfig)

// ExportDescription sets the export task description.
func ExportDescription(description string) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Description = description
	}
}

// ExportToGCS exports to Google Cloud Storage.
func ExportToGCS(bucket, prefix string) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Destination = ExportToCloudStorage
		cfg.Bucket = bucket
		cfg.Prefix = prefix
	}
}

// ExportToGoogleDrive exports to Google Drive.
func ExportToGoogleDrive(folder string) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Destination = ExportToDrive
		cfg.Folder = folder
	}
}

// ExportToEEAsset exports to Earth Engine Asset.
func ExportToEEAsset(assetID string) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Destination = ExportToAsset
		cfg.AssetID = assetID
	}
}

// ExportScale sets the export scale in meters per pixel.
func ExportScale(meters float64) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Scale = meters
	}
}

// ExportCRS sets the coordinate reference system.
func ExportCRS(crs string) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.CRS = crs
	}
}

// ExportRegion sets the region to export.
func ExportRegion(region *earthengine.Geometry) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Region = region
	}
}

// ExportMaxPixels sets the maximum number of pixels to export.
func ExportMaxPixels(maxPixels int64) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.MaxPixels = maxPixels
	}
}

// ExportFormat sets the export format.
func ExportFileFormat(format ExportFormat) ExportImageOption {
	return func(cfg *ExportConfig) {
		cfg.Format = format
	}
}

// ExportImage exports an image to Cloud Storage, Drive, or Assets.
//
// Note: This creates an export task configuration. Full implementation requires:
// 1. Earth Engine export API support
// 2. Asynchronous task handling
// 3. Progress tracking
// 4. Task status polling
//
// Current implementation provides the structure and validation but cannot
// execute actual exports. Use Earth Engine Code Editor or Python API for exports.
//
// Example:
//
//	// Configure an export (structure only - doesn't execute)
//	config := &ExportConfig{}
//	ExportDescription("My Export")(config)
//	ExportToGCS("my-bucket", "exports/")(config)
//	ExportScale(30)(config)
//
//	// Would need async task API to execute:
//	// task, err := helpers.ExportImage(ctx, client, image, config)
func ExportImage(ctx context.Context, client *earthengine.Client, image *earthengine.Image, opts ...ExportImageOption) error {
	// Apply options
	cfg := &ExportConfig{
		Description: "Export",
		Destination: ExportToCloudStorage,
		Format:      GeoTIFF,
		Scale:       30,
		CRS:         "EPSG:4326",
		MaxPixels:   1e9,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate configuration
	if err := validateExportConfig(cfg); err != nil {
		return err
	}

	// Note: Full implementation would:
	// 1. Build export task request
	// 2. Submit to Earth Engine export API
	// 3. Return task handle for progress tracking
	// 4. Support async completion notification
	//
	// This requires Earth Engine export REST API which is not yet implemented.

	return fmt.Errorf("export functionality requires Earth Engine export API support (not yet implemented)\n\n" +
		"Export configuration validated successfully:\n" +
		"  Description: %s\n" +
		"  Destination: %s\n" +
		"  Format: %s\n" +
		"  Scale: %.0fm\n" +
		"  CRS: %s\n\n" +
		"To export this image, use:\n" +
		"  - Earth Engine Code Editor (code.earthengine.google.com)\n" +
		"  - Earth Engine Python API\n" +
		"  - Or wait for go-earthengine export API implementation",
		cfg.Description, cfg.Destination, cfg.Format, cfg.Scale, cfg.CRS)
}

// validateExportConfig validates an export configuration.
func validateExportConfig(cfg *ExportConfig) error {
	if cfg.Description == "" {
		return fmt.Errorf("export description is required")
	}

	switch cfg.Destination {
	case ExportToCloudStorage:
		if cfg.Bucket == "" {
			return fmt.Errorf("bucket is required for Cloud Storage exports")
		}
	case ExportToDrive:
		// Folder is optional for Drive
	case ExportToAsset:
		if cfg.AssetID == "" {
			return fmt.Errorf("asset ID is required for Asset exports")
		}
	default:
		return fmt.Errorf("unsupported export destination: %s", cfg.Destination)
	}

	if cfg.Scale <= 0 {
		return fmt.Errorf("scale must be positive, got %.2f", cfg.Scale)
	}

	if cfg.MaxPixels <= 0 {
		return fmt.Errorf("maxPixels must be positive, got %d", cfg.MaxPixels)
	}

	return nil
}

// ExportTable exports a feature collection to Cloud Storage or Drive.
//
// Note: Placeholder - requires export API support.
func ExportTable(ctx context.Context, client *earthengine.Client, collection interface{}, opts ...ExportImageOption) error {
	cfg := &ExportConfig{
		Description: "Table Export",
		Destination: ExportToCloudStorage,
		Format:      CSV,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := validateExportConfig(cfg); err != nil {
		return err
	}

	return fmt.Errorf("table export requires Earth Engine export API support (not yet implemented)")
}

// ExportVideo exports a time series as a video.
//
// Note: Placeholder - requires export API support.
func ExportVideo(ctx context.Context, client *earthengine.Client, collection *earthengine.ImageCollection, opts ...ExportImageOption) error {
	cfg := &ExportConfig{
		Description: "Video Export",
		Destination: ExportToCloudStorage,
		Format:      MP4,
		Scale:       1000, // Default to 1km for video
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := validateExportConfig(cfg); err != nil {
		return err
	}

	return fmt.Errorf("video export requires Earth Engine export API support (not yet implemented)")
}
