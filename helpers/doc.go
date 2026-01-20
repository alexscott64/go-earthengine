// Package helpers provides high-level, domain-specific helpers for common Earth Engine tasks.
//
// This package makes it easy to work with Earth Engine data without dealing with
// low-level expression graphs. Each helper focuses on a specific domain:
//
// # Land Cover
//
// Query land cover data from NLCD, ESA WorldCover, and other datasets:
//
//	coverage, err := helpers.TreeCoverage(client, lat, lon, helpers.Latest())
//	landcover, err := helpers.LandCoverClass(client, lat, lon, "2020")
//
// # Elevation
//
// Get elevation data from SRTM, ASTER, and other DEMs:
//
//	elevation, err := helpers.Elevation(client, lat, lon, helpers.SRTM())
//	slope, err := helpers.Slope(client, lat, lon)
//
// # Imagery
//
// Work with Landsat, Sentinel, MODIS, and other imagery:
//
//	ndvi, err := helpers.NDVI(client, lat, lon, date, helpers.Landsat8())
//	composite, err := helpers.Composite(client, bounds, startDate, endDate, helpers.Sentinel2())
//
// # Solar & Astronomical
//
// Calculate sun angles, day length, and more:
//
//	sunAngle, err := helpers.SunAngle(lat, lon, time.Now())
//	dayLength, err := helpers.DayLength(lat, date)
//
// # Export
//
// Export data with progress tracking:
//
//	export := helpers.Export(client, image).
//	    ToCloudStorage("bucket", "prefix").
//	    WithProgress(func(pct float64) { fmt.Printf("%.1f%%\n", pct*100) })
//	err := export.Wait(ctx)
//
// # Batch Operations
//
// Process multiple queries in parallel:
//
//	batch := helpers.NewBatch(client, 10) // 10 concurrent
//	for _, location := range locations {
//	    batch.Add(helpers.TreeCoverageQuery(location.Lat, location.Lon))
//	}
//	results, err := batch.Execute(ctx)
//
// # Design Philosophy
//
// Helpers are designed to be:
//   - Simple: One function call for common tasks
//   - Flexible: Options for customization
//   - Composable: Combine multiple helpers
//   - Fast: Batch operations built-in
//   - Safe: Type-safe with clear errors
package helpers
