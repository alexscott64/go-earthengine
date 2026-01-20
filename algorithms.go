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
	AlgorithmImageCollectionFilterDate     = "ImageCollection.filterDate"
	AlgorithmImageCollectionReduce         = "ImageCollection.reduce"
	AlgorithmImageCollectionCount          = "ImageCollection.count"

	// Image math algorithms
	AlgorithmImageAdd              = "Image.add"
	AlgorithmImageSubtract         = "Image.subtract"
	AlgorithmImageMultiply         = "Image.multiply"
	AlgorithmImageDivide           = "Image.divide"
	AlgorithmImageNormalizedDiff   = "Image.normalizedDifference"
	AlgorithmImageExpression       = "Image.expression"

	// Date constructors
	AlgorithmDate = "Date"

	// Geometry constructors
	AlgorithmGeometryPoint = "GeometryConstructors.Point"

	// Reducer algorithms
	AlgorithmReducerFirst  = "Reducer.first"
	AlgorithmReducerMean   = "Reducer.mean"
	AlgorithmReducerMedian = "Reducer.median"
	AlgorithmReducerSum    = "Reducer.sum"
	AlgorithmReducerMin    = "Reducer.min"
	AlgorithmReducerMax    = "Reducer.max"
	AlgorithmReducerCount  = "Reducer.count"

	// Terrain algorithms
	AlgorithmTerrainSlope  = "Terrain.slope"
	AlgorithmTerrainAspect = "Terrain.aspect"
)
