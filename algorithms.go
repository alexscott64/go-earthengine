package earthengine

// Algorithm name constants for Earth Engine REST API.
// Note: Function names should NOT include the "algorithms/" prefix.
const (
	// Image algorithms
	AlgorithmImageLoad         = "Image.load"
	AlgorithmImageSelect       = "Image.select"
	AlgorithmImageReduceRegion = "Image.reduceRegion"

	// ImageCollection algorithms
	AlgorithmImageCollectionLoad           = "ImageCollection.load"
	AlgorithmImageCollectionFirst          = "ImageCollection.first"
	AlgorithmImageCollectionMosaic         = "ImageCollection.mosaic"
	AlgorithmImageCollectionFilterMetadata = "ImageCollection.filterMetadata"

	// Geometry constructors
	AlgorithmGeometryPoint = "GeometryConstructors.Point"

	// Reducer algorithms
	AlgorithmReducerFirst = "Reducer.first"
	AlgorithmReducerMean  = "Reducer.mean"
	AlgorithmReducerSum   = "Reducer.sum"
	AlgorithmReducerMin   = "Reducer.min"
	AlgorithmReducerMax   = "Reducer.max"
	AlgorithmReducerCount = "Reducer.count"
)
