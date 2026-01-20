package earthengine

import "fmt"

// ImageCollection represents an Earth Engine ImageCollection with chainable operations.
type ImageCollection struct {
	client       *Client
	expr         *ExpressionBuilder
	collectionID string // Collection ID for the ImageCollection
	nodeID       string // Current node ID in the expression graph (for chained operations)
}

// ImageCollection creates a new ImageCollection from an Earth Engine collection.
func (c *Client) ImageCollection(collectionID string) *ImageCollection {
	expr := NewExpressionBuilder()

	// Create ImageCollection.load node
	loadNodeID := expr.FunctionCall(AlgorithmImageCollectionLoad, map[string]interface{}{
		"id": map[string]interface{}{
			"constantValue": collectionID,
		},
	})

	return &ImageCollection{
		client:       c,
		expr:         expr,
		collectionID: collectionID,
		nodeID:       loadNodeID,
	}
}

// First returns the first image from the collection as an Image.
// This is useful for collections sorted by date or other criteria.
func (ic *ImageCollection) First() *Image {
	// Create Collection.first node to get the first image
	firstNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionFirst, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: firstNodeID,
	}
}

// FilterDate filters the collection to images within a date range.
// Dates should be in ISO 8601 format: "YYYY-MM-DD"
//
// Example:
//
//	collection := client.ImageCollection("COPERNICUS/S2_SR_HARMONIZED").
//	    FilterDate("2023-06-01", "2023-06-30")
func (ic *ImageCollection) FilterDate(startDate, endDate string) *ImageCollection {
	// Create Date objects for start and end
	startDateNode := ic.expr.FunctionCall(AlgorithmDate, map[string]interface{}{
		"value": map[string]interface{}{
			"constantValue": startDate,
		},
	})

	endDateNode := ic.expr.FunctionCall(AlgorithmDate, map[string]interface{}{
		"value": map[string]interface{}{
			"constantValue": endDate,
		},
	})

	// Create FilterDate node
	filterNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionFilterDate, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
		"start": map[string]interface{}{
			"valueReference": startDateNode,
		},
		"end": map[string]interface{}{
			"valueReference": endDateNode,
		},
	})

	// Return new ImageCollection with updated node
	return &ImageCollection{
		client:       ic.client,
		expr:         ic.expr,
		collectionID: ic.collectionID,
		nodeID:       filterNodeID,
	}
}

// FilterMetadata filters the collection based on metadata properties.
//
// Example:
//
//	collection := client.ImageCollection("COPERNICUS/S2_SR_HARMONIZED").
//	    FilterMetadata("CLOUDY_PIXEL_PERCENTAGE", "less_than", 20)
func (ic *ImageCollection) FilterMetadata(property string, operator string, value interface{}) *ImageCollection {
	filterNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionFilterMetadata, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
		"property": map[string]interface{}{
			"constantValue": property,
		},
		"operator": map[string]interface{}{
			"constantValue": operator,
		},
		"value": map[string]interface{}{
			"constantValue": value,
		},
	})

	return &ImageCollection{
		client:       ic.client,
		expr:         ic.expr,
		collectionID: ic.collectionID,
		nodeID:       filterNodeID,
	}
}

// FilterByYear filters the collection to images from a specific year.
// This is a convenience method for NLCD and other annual datasets.
func (ic *ImageCollection) FilterByYear(year int) *ImageCollection {
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)
	return ic.FilterDate(startDate, endDate)
}

// Reduce reduces the collection to a single image using a reducer.
//
// Example:
//
//	meanImage := collection.Reduce(earthengine.ReducerMean())
func (ic *ImageCollection) Reduce(reducer Reducer) *Image {
	// Get the reducer's expression representation
	reducerNodeID := reducer.NodeID(ic.expr)

	// Create reduce node
	reduceNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionReduce, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
		"reducer": map[string]interface{}{
			"valueReference": reducerNodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: reduceNodeID,
	}
}

// Count returns an image containing the number of images in the collection at each pixel.
func (ic *ImageCollection) Count() *Image {
	countNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionCount, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: countNodeID,
	}
}

// Select selects specific bands from all images in the collection.
func (ic *ImageCollection) Select(bands ...string) *ImageCollection {
	selectNodeID := ic.expr.FunctionCall("ImageCollection.select", map[string]interface{}{
		"input": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
		"bandSelectors": map[string]interface{}{
			"constantValue": bands,
		},
	})

	return &ImageCollection{
		client:       ic.client,
		expr:         ic.expr,
		collectionID: ic.collectionID,
		nodeID:       selectNodeID,
	}
}

// Mosaic creates a composite image from the collection by mosaicking.
// Later images are rendered on top of earlier images.
func (ic *ImageCollection) Mosaic() *Image {
	// Create ImageCollection.mosaic node
	mosaicNodeID := ic.expr.FunctionCall(AlgorithmImageCollectionMosaic, map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": ic.nodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: mosaicNodeID,
	}
}
