# go-earthengine

A production-grade Go client library for Google Earth Engine REST API with high-level domain-specific helpers.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/)
[![Tests](https://img.shields.io/badge/tests-260%20passing-brightgreen)](https://github.com/alexscott64/go-earthengine)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Features

- ✅ **Complete REST API Client** - Full access to Earth Engine REST API v1
- ✅ **High-Level Helpers** - Domain-specific convenience methods
- ✅ **Batch Operations** - Parallel processing with concurrency control
- ✅ **Checkpoint Resume** - Automatic save/resume for long-running operations
- ✅ **Quota Tracking** - Monitor and limit daily API usage
- ✅ **Built-in Caching** - In-memory cache for improved performance
- ✅ **Async Export API** - Task submission, monitoring, progress tracking
- ✅ **Real Datasets** - NLCD 2023, Hansen GFC, SRTM, Landsat, Sentinel-2
- ✅ **Solar Calculations** - Sun position, sunrise/sunset, day length
- ✅ **Type-Safe** - Idiomatic Go with comprehensive error handling
- ✅ **Well Tested** - 260 tests with excellent coverage
- ✅ **Real-World Examples** - Tree coverage, slope analysis, sun position, exports

## Installation

```bash
go get github.com/alexscott64/go-earthengine
```

## Quick Start

### Setup

```go
import (
    "context"
    "fmt"
    "github.com/alexscott64/go-earthengine"
    "github.com/alexscott64/go-earthengine/helpers"
)

// Initialize client with service account
client, err := earthengine.NewClient(
    context.Background(),
    earthengine.WithProject("your-project-id"),
    earthengine.WithServiceAccountFile("path/to/service-account.json"),
)
if err != nil {
    panic(err)
}
```

### Simple Queries

```go
// Get tree coverage at a location
coverage, err := helpers.TreeCoverage(client, 45.5152, -122.6784)
fmt.Printf("Tree coverage: %.1f%%\n", coverage)

// Get elevation
elevation, err := helpers.Elevation(client, 39.7392, -104.9903)
fmt.Printf("Elevation: %.0f meters\n", elevation)

// Check if location is urban
urban, err := helpers.IsUrban(client, 45.5152, -122.6784)
if urban {
    fmt.Println("This is an urban area")
}

// Calculate sunrise and sunset
sunrise, _ := helpers.SunriseTime(45.5152, -122.6784, time.Now())
sunset, _ := helpers.SunsetTime(45.5152, -122.6784, time.Now())
fmt.Printf("Sunrise: %s, Sunset: %s\n",
    sunrise.Format("15:04"), sunset.Format("15:04"))
```

### Using Options

```go
// Get tree coverage from a specific year
coverage, err := helpers.TreeCoverage(client, 45.5152, -122.6784,
    helpers.Year(2020))

// Get high-resolution elevation (USA only)
elevation, err := helpers.Elevation(client, 39.7392, -104.9903,
    helpers.USGS3DEP(),              // 10m resolution
    helpers.ElevationWithScale(10))

// Use global dataset for non-USA locations
coverage, err := helpers.TreeCoverage(client, 52.5200, 13.4050, // Berlin
    helpers.HansenDataset())
```

### Batch Processing

Process multiple locations in parallel with automatic concurrency control:

```go
// Create batch with 10 concurrent queries
batch := helpers.NewBatch(client, 10)

// Add queries
locations := []struct{ Lat, Lon float64 }{
    {45.5152, -122.6784}, // Portland
    {47.6062, -122.3321}, // Seattle
    {37.7749, -122.4194}, // San Francisco
}

for _, loc := range locations {
    batch.Add(helpers.NewTreeCoverageQuery(loc.Lat, loc.Lon))
}

// Execute with progress tracking
results, err := batch.ExecuteWithProgress(ctx, func(completed, total int) {
    fmt.Printf("Progress: %d/%d (%.1f%%)\n",
        completed, total, float64(completed)*100/float64(total))
})

// Summarize results
summary := helpers.Summarize(results)
fmt.Printf("Success rate: %.1f%% (%d/%d)\n",
    summary.SuccessRate*100, summary.Succeeded, summary.Total)

// Process successful results
for _, result := range helpers.FilterSuccessful(results) {
    coverage := result.Value.(float64)
    fmt.Printf("Coverage at location %d: %.1f%%\n", result.Index, coverage)
}
```

### Advanced Batch Operations

```go
// Execute with automatic retry
results, err := batch.ExecuteWithRetry(ctx,
    3,                      // max retries
    100*time.Millisecond)   // initial backoff

// Rate limiting
limiter := helpers.NewRateLimiter(10) // 10 requests/second
defer limiter.Close()

for _, loc := range locations {
    limiter.Wait(ctx)
    result, err := helpers.TreeCoverage(client, loc.Lat, loc.Lon)
    // process result...
}
```

### Resume from Checkpoint

For long-running batch operations, automatically save and resume progress:

```go
// Create batch
batch := helpers.NewBatch(client, 10)
for _, loc := range locations {
    batch.Add(helpers.NewTreeCoverageQuery(loc.Lat, loc.Lon))
}

// Execute with checkpoint support
// If interrupted, will resume from where it left off
results, err := helpers.ExecuteWithCheckpoint(ctx, batch, "progress.json", 10)

// Clean up checkpoint after success
if err == nil {
    helpers.RemoveCheckpoint("progress.json")
}

// Check progress of existing checkpoint
completed, total, _ := helpers.GetCheckpointProgress("progress.json")
fmt.Printf("Progress: %d/%d (%.1f%%)\n",
    completed, total, float64(completed)*100/float64(total))
```

### Quota Tracking

Monitor and limit API usage:

```go
// Create quota tracker with daily limit
tracker := earthengine.NewQuotaTracker(1000) // 1000 requests/day

// Record each request
tracker.RecordRequest()

// Check quota status
if tracker.IsQuotaExceeded() {
    log.Fatalf("Daily quota exceeded")
}

// Get usage statistics
stats := tracker.GetUsageStats()
fmt.Printf("Today: %d requests\n", stats.TodayRequests)
fmt.Printf("Remaining: %d requests\n", stats.RemainingQuota)
fmt.Printf("Average: %.1f requests/hour\n", stats.RequestsPerHour)

// Cleanup old data (keep last 30 days)
tracker.CleanupOldData(30)
```

### Caching

Cache Earth Engine query results for better performance:

```go
// Create in-memory cache (max 1000 entries)
cache := earthengine.NewMemoryCache(1000)

// Manual caching
cacheKey := earthengine.CacheKey(lat, lon, date, "ndvi")
if cached, found, _ := cache.Get(ctx, cacheKey); found {
    result = cached.(float64)
} else {
    result, err = helpers.NDVI(client, lat, lon, date)
    if err == nil {
        cache.Set(ctx, cacheKey, result, 1*time.Hour)
    }
}

// Use cached client wrapper
cachedClient := earthengine.NewCachedClient(client, cache, 1*time.Hour)

// Cache statistics
stats := cache.Stats()
fmt.Printf("Cache size: %d/%d entries\n", stats.Size, stats.MaxSize)

// Clear cache
cache.Clear(ctx)
```

### Async Export API

Export Earth Engine data to Cloud Storage, Google Drive, or as Earth Engine Assets with full async task monitoring:

```go
// Simple image export
task, err := helpers.ExportImageAsync(ctx, client, image,
    helpers.ExportDescription("Summer 2023 NDVI"),
    helpers.ExportToGCS("my-bucket", "exports/"),
    helpers.ExportScale(30),
    helpers.ExportCRS("EPSG:4326"),
    helpers.ExportFileFormat(helpers.GeoTIFF))

if err != nil {
    log.Fatalf("Export failed: %v", err)
}

fmt.Printf("Export started: %s\n", task.ID)

// Wait with progress tracking
err = task.WaitWithProgress(ctx, func(progress *earthengine.TaskProgress) {
    fmt.Printf("\rProgress: %.1f%% | State: %s",
        progress.Progress*100,
        progress.State)
})

if err != nil {
    log.Fatalf("Export failed: %v", err)
}

fmt.Println("\nExport completed successfully!")

// Export to different destinations
taskGCS, _ := helpers.ExportImageAsync(ctx, client, image,
    helpers.ExportDescription("Export to GCS"),
    helpers.ExportToGCS("my-bucket", "exports/"))

taskDrive, _ := helpers.ExportImageAsync(ctx, client, image,
    helpers.ExportDescription("Export to Drive"),
    helpers.ExportToGoogleDrive("Earth Engine Exports"))

taskAsset, _ := helpers.ExportImageAsync(ctx, client, image,
    helpers.ExportDescription("Export as Asset"),
    helpers.ExportToEEAsset("projects/my-project/assets/my-image"))

// Export with notification callback
task, err := helpers.ExportImageWithNotification(ctx, client, image,
    func(t *earthengine.Task, err error) {
        if err != nil {
            log.Printf("❌ Export failed: %v", err)
            // Send email, trigger webhook, etc.
        } else {
            log.Printf("✓ Export completed: %s", t.ID)
        }
    },
    helpers.ExportDescription("Export with Notification"),
    helpers.ExportToGCS("my-bucket", "exports/"))

// Batch exports with progress monitoring
var tasks []*earthengine.Task
for _, region := range regions {
    task, err := helpers.ExportImageAsync(ctx, client, image,
        helpers.ExportDescription(fmt.Sprintf("Export %s", region.name)),
        helpers.ExportToGCS("my-bucket", fmt.Sprintf("regions/%s/", region.name)))

    if err == nil {
        tasks = append(tasks, task)
    }
}

// Wait for all exports
results := helpers.WaitForExports(ctx, tasks, func(completed, total int) {
    fmt.Printf("\rExports completed: %d/%d (%.1f%%)",
        completed, total, float64(completed)*100/float64(total))
})

// Check results
successful := 0
for i, result := range results {
    if result.Error != nil {
        log.Printf("❌ Export %d failed: %v\n", i+1, result.Error)
    } else {
        log.Printf("✓ Export %d completed: %s\n", i+1, result.TaskID)
        successful++
    }
}

fmt.Printf("\nSummary: %d/%d exports successful\n", successful, len(results))

// Task management
tm := earthengine.NewTaskManager(client)

// Register tasks
tm.RegisterTask(task)

// List all tasks
allTasks := tm.ListTasks()

// Filter active tasks
activeTasks := tm.FilterTasks(earthengine.TaskFilter{
    States: []earthengine.TaskState{
        earthengine.TaskStatePending,
        earthengine.TaskStateRunning,
    },
})

// Cancel all active tasks
tm.CancelAll(ctx)

// Cleanup old tasks (remove completed tasks older than 24 hours)
tm.Cleanup(24 * time.Hour)

// Export other data types
tableTask, _ := helpers.ExportTableAsync(ctx, client, collection,
    helpers.ExportDescription("Feature Collection Export"),
    helpers.ExportToGCS("my-bucket", "tables/"),
    helpers.ExportFileFormat(helpers.CSV))

videoTask, _ := helpers.ExportVideoAsync(ctx, client, imageCollection,
    helpers.ExportDescription("Time Lapse Video"),
    helpers.ExportToGCS("my-bucket", "videos/"),
    helpers.ExportFileFormat(helpers.MP4))
```

## Domain Helpers

### Land Cover

```go
// Tree canopy coverage (NLCD 2023 or Hansen Global Forest Change)
coverage, err := helpers.TreeCoverage(client, lat, lon)

// Land cover classification
class, err := helpers.LandCoverClass(client, lat, lon)
// Returns: "forest_evergreen", "developed_medium", "water", etc.

// Impervious surface percentage
impervious, err := helpers.ImperviousSurface(client, lat, lon)

// Urban detection
isUrban, err := helpers.IsUrban(client, lat, lon)
```

**Datasets**: NLCD 2023 (USA), Hansen Global Forest Change 2023 (global)

### Elevation

```go
// Get elevation with default dataset (SRTM 30m)
elevation, err := helpers.Elevation(client, lat, lon)

// Choose specific dataset
elevation, err := helpers.Elevation(client, lat, lon, helpers.USGS3DEP())

// Get comprehensive terrain metrics
metrics, err := helpers.TerrainAnalysis(client, lat, lon)
fmt.Printf("Elevation: %.0fm, Slope: %.1f°, Aspect: %.0f°\n",
    metrics.Elevation, metrics.Slope, metrics.Aspect)
```

**Datasets**: SRTM 30m, ASTER 30m, ALOS 30m, USGS 3DEP 10m (USA)

### Geometry

```go
// Calculate distance between two points
distance := helpers.DistanceMeters(
    45.5152, -122.6784,  // Portland
    47.6062, -122.3321,  // Seattle
)
fmt.Printf("Distance: %.0f km\n", distance/1000)

// Create and manipulate bounds
bounds := helpers.BoundsFromPoints([][2]float64{
    {-122.5, 45.4}, {-122.3, 45.6},
})
centerLat, centerLon := bounds.Center()
expanded := bounds.Expand(0.1) // 10% larger

// Check if point is within bounds
if bounds.Contains(lat, lon) {
    fmt.Println("Point is within bounds")
}
```

### Solar/Astronomical

```go
// Sun position
pos, err := helpers.CalculateSunPosition(lat, lon, time.Now())
fmt.Printf("Sun: Azimuth %.0f°, Elevation %.0f°\n",
    pos.Azimuth, pos.Elevation)

// Daylight hours
date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
dayLength, err := helpers.DayLength(lat, date)
fmt.Printf("Daylight: %.1f hours\n", dayLength.Hours())

// Sunrise and sunset
sunrise, err := helpers.SunriseTime(lat, lon, date)
sunset, err := helpers.SunsetTime(lat, lon, date)

// Check if it's daytime
isDaytime, err := helpers.IsDaytime(lat, lon, time.Now())

// Solar noon
solarNoon, err := helpers.SolarNoon(lon, date)
```

**Features**: Accurate calculations, handles polar day/night, UTC times

### Imagery (Structure Complete)

```go
// Vegetation indices (requires image band math support)
ndvi, err := helpers.NDVI(client, lat, lon, "2023-06-01",
    helpers.Sentinel2(),
    helpers.CloudMask(20))

// Other indices: EVI, SAVI, NDWI, NDBI
// Spectral bands retrieval
// Composite creation
```

**Datasets**: Landsat 8/9, Sentinel-2, MODIS

## Low-Level API

For advanced use cases, use the low-level API client directly:

```go
import "github.com/alexscott64/go-earthengine/apiv1"

// Create service
svc, err := apiv1.NewService(ctx,
    apiv1.WithServiceAccountFile("service-account.json"))

// Compute a value
result, err := svc.Projects.Value.Compute(ctx, "projects/your-project",
    &apiv1.ComputeValueRequest{
        Expression: expression,
    })

// Export an image
op, err := svc.Projects.Image.Export(ctx, "projects/your-project",
    &apiv1.ExportImageRequest{
        Expression:    imageExpression,
        Description:   "my-export",
        FileExportOptions: &apiv1.ImageFileExportOptions{
            // configuration...
        },
    })

// Wait for operation
completed, err := svc.Projects.Operations.WaitWithPolling(ctx,
    op.Name, 5*time.Second)
```

## Architecture

```
User Code
    ↓
helpers/        ← High-level domain helpers (TreeCoverage, Elevation, etc.)
    ↓
client.go       ← Mid-level client (Image, ImageCollection operations)
    ↓
apiv1/          ← Low-level API client (complete REST API access)
    ↓
Earth Engine REST API
```

## Authentication

### Service Account JSON File

```go
client, err := earthengine.NewClient(ctx,
    earthengine.WithProject("your-project-id"),
    earthengine.WithServiceAccountFile("service-account.json"))
```

### Service Account JSON Data

```go
jsonData, _ := os.ReadFile("service-account.json")
client, err := earthengine.NewClient(ctx,
    earthengine.WithProject("your-project-id"),
    earthengine.WithServiceAccountJSON(jsonData))
```

### Environment Variables

```bash
export GOOGLE_EARTH_ENGINE_PROJECT_ID="your-project-id"
export GOOGLE_EARTH_ENGINE_CLIENT_EMAIL="service@project.iam.gserviceaccount.com"
export GOOGLE_EARTH_ENGINE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n..."
```

```go
client, err := earthengine.NewClient(ctx,
    earthengine.WithServiceAccountEnv())
```

## Error Handling

All functions return standard Go errors:

```go
coverage, err := helpers.TreeCoverage(client, lat, lon)
if err != nil {
    // Handle error
    log.Printf("Failed to get tree coverage: %v", err)
    return
}
```

For batch operations, errors are per-query:

```go
results, err := batch.Execute(ctx)
if err != nil {
    // Fatal error (e.g., context canceled)
    return err
}

for i, result := range results {
    if result.Error != nil {
        log.Printf("Query %d failed: %v", i, result.Error)
        continue
    }
    // Process result.Value
}
```

## Context Support

All operations support context for cancellation and timeouts:

```go
// Timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := helpers.TreeCoverageWithContext(ctx, client, lat, lon)

// Cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Cancel after some condition
    cancel()
}()

results, err := batch.Execute(ctx)
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./helpers -v
go test ./apiv1 -v

# Run with coverage
go test ./... -cover
```

**Current Status**: 260 tests, all passing ✅

## Datasets

### Integrated Datasets

- **NLCD 2023** - Land cover, tree canopy, impervious surface (USA, 30m)
- **Hansen GFC 2023** - Global forest change (global, 30m)
- **SRTM** - Elevation (near-global, 30m)
- **ASTER GDEM** - Elevation (global, 30m)
- **ALOS World 3D** - Elevation (global, 30m)
- **USGS 3DEP** - High-res elevation (USA, 10m)
- **Landsat 8/9** - Multispectral imagery (global, 30m)
- **Sentinel-2** - Multispectral imagery (global, 10-20m)
- **MODIS** - Vegetation indices (global, 250m-1km)

## Performance

- **Parallel batch processing** with configurable concurrency
- **Automatic retry** with exponential backoff
- **Rate limiting** to respect API quotas
- **Context cancellation** for graceful shutdown
- **Efficient caching** where applicable

## Roadmap

### Completed ✅
- Low-level API client (apiv1)
- Land cover helpers
- Elevation helpers
- Geometry helpers
- Batch operations
- Solar/astronomical helpers
- Image band math support (Add, Subtract, Multiply, Divide, NormalizedDifference, Expression)
- ImageCollection filtering (FilterDate, FilterMetadata, Reduce, Count, Select)
- Climate helpers (Temperature, Precipitation, SoilMoisture)
- Imagery helpers (NDVI, EVI, SAVI, NDWI, NDBI, SpectralBands, Composite)
- Water helpers (WaterDetection, WaterOccurrence, WaterSeasonality, WaterChange)
- Fire helpers (ActiveFire, FireCount, BurnSeverity, DeltaNBR)
- Terrain algorithms (Slope, Aspect)
- Export helpers (ExportImage, ExportTable, ExportVideo) - Configuration only
- Export API Integration with async task support:
  - Task submission and monitoring
  - Progress tracking for long-running exports
  - Async completion notifications
  - Batch export monitoring with WaitForExports
  - Task management (filter, cancel, cleanup)

### Planned 📋

**Advanced Features**:
- Time-series analysis and trend detection
- Advanced compositing methods (median, quality mosaics)
- Zonal statistics over polygons

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Google Earth Engine team for the excellent API
- Earth Engine dataset providers (USGS, NASA, ESA, etc.)
- Go community for best practices and patterns

---

**Built with ❤️ for the Earth Engine community**
