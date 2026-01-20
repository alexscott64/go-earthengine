# Quick Start Guide

## Installation

```bash
go get github.com/yourusername/go-earthengine
```

## Basic Usage

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

    // Create client
    client, err := ee.NewClient(ctx,
        ee.WithServiceAccountFile("credentials.json"),
        ee.WithProject("your-project-id"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Quick tree coverage query
    coverage, err := client.GetTreeCoverage(ctx, 47.6, -120.9)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Tree coverage: %.2f%%\n", coverage)
}
```

## Using the Fluent API

```go
// Query any dataset
result, err := client.Image("USGS/NLCD/NLCD2016").
    Select("percent_tree_cover").
    ReduceRegion(
        ee.NewPoint(-120.9, 47.6),
        ee.ReducerFirst(),
        ee.Scale(30),
    ).
    Compute(ctx)

fmt.Printf("Result: %v\n", result)
```

## Authentication Options

### Option 1: JSON File
```go
ee.WithServiceAccountFile("path/to/credentials.json")
```

### Option 2: Environment Variables
```bash
export GOOGLE_EARTH_ENGINE_PROJECT_ID=your-project
export GOOGLE_EARTH_ENGINE_CLIENT_EMAIL=your-email@project.iam.gserviceaccount.com
export GOOGLE_EARTH_ENGINE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n..."
```

```go
ee.WithServiceAccountEnv()
```

## Available Reducers

```go
ee.ReducerFirst()   // Get first value
ee.ReducerMean()    // Calculate mean
ee.ReducerMin()     // Find minimum
ee.ReducerMax()     // Find maximum
ee.ReducerSum()     // Sum all values
ee.ReducerCount()   // Count values
```

## Running Examples

```bash
# Set environment variables
export GOOGLE_EARTH_ENGINE_PROJECT_ID=your-project
export GOOGLE_EARTH_ENGINE_CLIENT_EMAIL=your-email
export GOOGLE_EARTH_ENGINE_PRIVATE_KEY="your-key"

# Run tree coverage example
cd examples/tree_coverage
go run main.go

# Run basic query example
cd examples/basic_query
go run main.go
```

## Running Tests

```bash
# Unit tests only
go test -v -short ./...

# All tests (requires credentials)
go test -v ./...
```

## Common Datasets

- `USGS/NLCD/NLCD2016` - National Land Cover Database (USA)
- `MODIS/006/MOD13Q1` - MODIS Vegetation Indices
- `COPERNICUS/S2` - Sentinel-2 Imagery

See [Earth Engine Data Catalog](https://developers.google.com/earth-engine/datasets) for more.
