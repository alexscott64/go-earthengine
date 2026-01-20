# go-earthengine

A simple, idiomatic Go library for Google Earth Engine's REST API. Query satellite imagery, land cover data, and other geospatial datasets with a clean, chainable API.

## Features

- 🔐 **Service Account Authentication** - Support for JSON key files and environment variables
- 🔗 **Fluent API** - Chainable methods for building Earth Engine queries
- 🌲 **Convenience Methods** - Quick helpers for common operations (tree coverage, etc.)
- 📦 **Type-Safe** - Uses Go types instead of raw JSON manipulation
- 🧪 **Well-Tested** - Comprehensive unit and integration tests
- 🚀 **Minimal Dependencies** - Just Go standard library and Google OAuth2

## Installation

```bash
go get github.com/yourusername/go-earthengine
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    ee "github.com/yourusername/go-earthengine"
)

func main() {
    ctx := context.Background()

    // Create client with service account authentication
    client, err := ee.NewClient(ctx,
        ee.WithServiceAccountFile("credentials.json"),
        ee.WithProject("my-gcp-project"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get tree coverage at a location (convenience method)
    coverage, err := client.GetTreeCoverage(ctx, 47.6, -120.9)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Tree coverage: %.2f%%\n", coverage)
}
```

## Authentication

### Option 1: Service Account JSON File

```go
client, err := ee.NewClient(ctx,
    ee.WithServiceAccountFile("path/to/credentials.json"),
    ee.WithProject("your-gcp-project-id"),
)
```

### Option 2: Environment Variables

Set these environment variables:
```bash
export GOOGLE_EARTH_ENGINE_PROJECT_ID=your-project-id
export GOOGLE_EARTH_ENGINE_CLIENT_EMAIL=your-service-account@project.iam.gserviceaccount.com
export GOOGLE_EARTH_ENGINE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n..."
```

Then create the client:
```go
client, err := ee.NewClient(ctx,
    ee.WithServiceAccountEnv(),
    ee.WithProject(os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")),
)
```

### Setting Up Google Earth Engine Access

1. **Create a Google Cloud Project** at https://console.cloud.google.com
2. **Enable Earth Engine API** in your project
3. **Create a Service Account**:
   - Go to IAM & Admin → Service Accounts
   - Create a new service account
   - Grant it the "Earth Engine Resource Viewer" role
   - Create and download a JSON key
4. **Register for Earth Engine** at https://signup.earthengine.google.com

## Usage Examples

### Basic Query with Fluent API

```go
// Query NLCD tree coverage at a specific point
result, err := client.Image("USGS/NLCD/NLCD2016").
    Select("percent_tree_cover").
    ReduceRegion(
        ee.NewPoint(-120.9, 47.6),  // longitude, latitude
        ee.ReducerFirst(),
        ee.Scale(30),  // 30 meter resolution
    ).
    Compute(ctx)

fmt.Printf("Tree coverage: %v\n", result["percent_tree_cover"])
```

### Multiple Bands

```go
// Query multiple bands at once
result, err := client.Image("USGS/NLCD/NLCD2016").
    Select("percent_tree_cover", "impervious").
    ReduceRegion(
        ee.NewPoint(-122.3321, 47.6062),
        ee.ReducerFirst(),
        ee.Scale(30),
    ).
    Compute(ctx)

fmt.Printf("Tree coverage: %v%%\n", result["percent_tree_cover"])
fmt.Printf("Impervious surface: %v%%\n", result["impervious"])
```

### Different Reducers

```go
// Use different reducers for different analysis
reducers := []struct {
    name    string
    reducer ee.Reducer
}{
    {"First", ee.ReducerFirst()},
    {"Mean", ee.ReducerMean()},
    {"Min", ee.ReducerMin()},
    {"Max", ee.ReducerMax()},
}

for _, r := range reducers {
    result, _ := client.Image("USGS/NLCD/NLCD2016").
        Select("percent_tree_cover").
        ReduceRegion(
            ee.NewPoint(-120.9, 47.6),
            r.reducer,
            ee.Scale(30),
        ).
        Compute(ctx)

    fmt.Printf("%s: %v\n", r.name, result)
}
```

### Convenience Helper Methods

```go
// Get tree coverage (uses NLCD 2016 dataset)
coverage, err := client.GetTreeCoverage(ctx, latitude, longitude)

// Get detailed tree coverage with metadata
result, err := client.GetTreeCoverageDetailed(ctx, latitude, longitude)
fmt.Printf("Coverage: %.2f%% from %s\n", result.Coverage, result.DataSource)
```

## API Design

The library uses a graph-based expression builder that mirrors Earth Engine's internal structure:

```go
// Each operation creates a node in the expression graph
client.Image("dataset-id")           // Node: Image.load
    .Select("band1", "band2")        // Node: Image.select
    .ReduceRegion(                   // Node: Image.reduceRegion
        ee.NewPoint(lon, lat),       // Node: GeometryConstructors.Point
        ee.ReducerFirst(),           // Node: Reducer.first
        ee.Scale(30),
    ).Compute(ctx)                   // Execute the expression
```

## Supported Datasets

The library works with any Earth Engine dataset. Here are some commonly used ones:

### Land Cover (USA)
- **USGS/NLCD/NLCD2016** - National Land Cover Database 2016
  - Bands: `percent_tree_cover`, `impervious`, `landcover`
  - Resolution: 30m
  - Coverage: Continental USA

### Global Datasets
- **MODIS** - Terra and Aqua satellite imagery
- **GEDI** - Global Ecosystem Dynamics Investigation (forest structure)
- **Sentinel-2** - High-resolution multispectral imagery
- **Landsat** - Landsat 4-9 satellite imagery

See the [Earth Engine Data Catalog](https://developers.google.com/earth-engine/datasets) for the full list.

## Error Handling

The library provides clear error messages:

```go
coverage, err := client.GetTreeCoverage(ctx, 47.6, -120.9)
if err != nil {
    // Errors include:
    // - Authentication failures
    // - Invalid coordinates
    // - API errors with status codes
    // - Network errors
    log.Fatalf("Error: %v", err)
}
```

## Testing

Run the unit tests:
```bash
go test -v -short ./...
```

Run integration tests (requires credentials):
```bash
export GOOGLE_EARTH_ENGINE_PROJECT_ID=your-project
export GOOGLE_EARTH_ENGINE_CLIENT_EMAIL=your-email
export GOOGLE_EARTH_ENGINE_PRIVATE_KEY="your-key"
go test -v ./...
```

## Examples

See the [`examples/`](examples/) directory for complete working examples:

- **[tree_coverage/](examples/tree_coverage/)** - Get tree canopy coverage at various locations
- **[basic_query/](examples/basic_query/)** - Basic queries with different reducers

Run an example:
```bash
cd examples/tree_coverage
go run main.go
```

## Limitations

- Currently supports only the `value:compute` endpoint (point queries)
- Image export and other advanced features planned for future releases
- NLCD datasets only cover the United States

## Roadmap

- [ ] Image export functionality
- [ ] Feature collection support
- [ ] Image collection filtering and mapping
- [ ] More convenience methods for popular datasets
- [ ] Batch processing support

## Contributing

Contributions welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built for the [Google Earth Engine REST API](https://developers.google.com/earth-engine/apidocs)
- Inspired by the Python `earthengine-api` library

## Resources

- [Earth Engine Data Catalog](https://developers.google.com/earth-engine/datasets)
- [Earth Engine API Documentation](https://developers.google.com/earth-engine/apidocs)
- [Google Cloud Service Accounts](https://cloud.google.com/iam/docs/service-accounts)
