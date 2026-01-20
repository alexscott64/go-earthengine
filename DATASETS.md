# Available Earth Engine Datasets

This document lists commonly used datasets and how to access them with go-earthengine.

## Tree Canopy Cover (USA)

### NLCD 2016 (Current Default)
**Dataset ID**: `USGS/NLCD/NLCD2016`
**Status**: ✅ Working with library
**Coverage**: Continental USA
**Year**: 2016
**Bands**: `percent_tree_cover`, `impervious`, `landcover`
**Resolution**: 30m

```go
coverage, err := client.GetTreeCoverage(ctx, latitude, longitude)
```

### NLCD 2023 Tree Canopy Cover (Latest)
**Dataset ID**: `USGS/NLCD_RELEASES/2023_REL/TCC/v2023-5`
**Status**: ⚠️ ImageCollection (requires filtering, not yet supported)
**Coverage**: Continental USA (Alaska, Hawaii, Puerto Rico coming 2025)
**Years**: Annual data 1985-2023
**Bands**: `NLCD_Percent_Tree_Canopy_Cover`, `Science_Percent_Tree_Canopy_Cover`
**Resolution**: 30m

**Documentation**: [Earth Engine Catalog](https://developers.google.com/earth-engine/datasets/catalog/USGS_NLCD_RELEASES_2023_REL_TCC_v2023-5)

To use this once ImageCollection support is added:
```go
// Future API (not yet implemented)
coverage, err := client.ImageCollection("USGS/NLCD_RELEASES/2023_REL/TCC/v2023-5").
    FilterDate("2023-01-01", "2023-12-31").
    First().
    Select("NLCD_Percent_Tree_Canopy_Cover").
    ReduceRegion(...)
```

### Annual NLCD (1985-2024)
**Dataset ID**: `projects/sat-io/open-datasets/USGS/ANNUAL_NLCD/LANDCOVER`
**Status**: ⚠️ Community catalog, ImageCollection
**Coverage**: Continental USA
**Years**: Annual data 1985-2024
**Resolution**: 30m

**Documentation**: [Community Catalog](https://gee-community-catalog.org/projects/annual_nlcd/)

## Global Datasets

### Hansen Global Forest Change
**Dataset ID**: `UMD/hansen/global_forest_change_2023_v1_11`
**Status**: ✅ Should work (single Image)
**Coverage**: Global
**Bands**: `treecover2000`, `loss`, `gain`, `lossyear`
**Resolution**: 30m

```go
result, err := client.Image("UMD/hansen/global_forest_change_2023_v1_11").
    Select("treecover2000").
    ReduceRegion(
        ee.NewPoint(lon, lat),
        ee.ReducerFirst(),
        ee.Scale(30),
    ).
    Compute(ctx)
```

### MODIS Vegetation Indices
**Dataset ID**: `MODIS/006/MOD13A2` (16-day)
**Status**: ⚠️ ImageCollection
**Coverage**: Global
**Bands**: `NDVI`, `EVI`
**Resolution**: 1000m

### Sentinel-2 Surface Reflectance
**Dataset ID**: `COPERNICUS/S2_SR_HARMONIZED`
**Status**: ⚠️ ImageCollection
**Coverage**: Global
**Bands**: Multiple spectral bands
**Resolution**: 10-60m

## How to Find Dataset IDs

1. **Earth Engine Data Catalog**: https://developers.google.com/earth-engine/datasets/
   - Browse by category (Land Cover, Vegetation, Climate, etc.)
   - Each dataset page shows the Earth Engine Snippet

2. **Community Catalog**: https://gee-community-catalog.org/
   - Additional datasets contributed by the community
   - Often includes newer or specialized datasets

3. **Check Dataset Type**:
   - **Image**: Single static image (works with current library)
   - **ImageCollection**: Time series of images (requires filtering, planned for future)

## Discovering Band Names

To find available bands in a dataset:

1. Check the Earth Engine Data Catalog documentation
2. Look at the "Bands" table on the dataset page
3. Use the Earth Engine Code Editor to inspect the dataset

Example from Code Editor:
```javascript
var image = ee.Image('USGS/NLCD/NLCD2016');
print('Bands:', image.bandNames());
```

## Upcoming Library Features

The following features are planned to support more datasets:

- [ ] **ImageCollection Support** - Filter and select from time series
- [ ] **Date Filtering** - Select images by date range
- [ ] **Cloud Masking** - Automatic cloud filtering for optical imagery
- [ ] **Mosaicking** - Combine multiple images
- [ ] **Collection Reduction** - Aggregate over time (mean, max, etc.)

## Recommended Datasets by Use Case

### Forest/Vegetation Monitoring (USA)
- **Current**: NLCD 2016 (`USGS/NLCD/NLCD2016`)
- **Latest**: NLCD 2023 TCC (once ImageCollection support is added)
- **Global**: Hansen Forest Change

### Land Cover Classification
- **USA**: NLCD 2016 (`USGS/NLCD/NLCD2016`)
- **Global**: ESA WorldCover, Dynamic World

### Elevation/Terrain
- **Global**: SRTM, ASTER GDEM, ALOS DEM
- **USA High-Res**: USGS 3DEP

### Climate/Weather
- **Temperature**: MODIS Land Surface Temperature
- **Precipitation**: GPM, CHIRPS
- **ERA5**: Comprehensive climate reanalysis

## Contributing

Know of a useful dataset? Please contribute to this documentation by:
1. Testing it with the library
2. Adding it to this list with working code examples
3. Submitting a pull request

## External Resources

- [Earth Engine Data Catalog](https://developers.google.com/earth-engine/datasets/)
- [Community Catalog](https://gee-community-catalog.org/)
- [Awesome GEE Community Datasets](https://github.com/samapriya/awesome-gee-community-datasets)
