package earthengine

// ImageCollection represents an Earth Engine ImageCollection with chainable operations.
type ImageCollection struct {
	client       *Client
	expr         *ExpressionBuilder
	collectionID string // Collection ID for the ImageCollection
}

// ImageCollection creates a new ImageCollection from an Earth Engine collection.
func (c *Client) ImageCollection(collectionID string) *ImageCollection {
	expr := NewExpressionBuilder()

	return &ImageCollection{
		client:       c,
		expr:         expr,
		collectionID: collectionID,
	}
}

// First returns the first image from the collection as an Image.
// This is useful for collections sorted by date or other criteria.
func (ic *ImageCollection) First() *Image {
	// Create ImageCollection.load node
	loadNodeID := ic.expr.FunctionCall("ImageCollection.load", map[string]interface{}{
		"id": map[string]interface{}{
			"constantValue": ic.collectionID,
		},
	})

	// Create Collection.first node to get the first image
	firstNodeID := ic.expr.FunctionCall("Collection.first", map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": loadNodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: firstNodeID,
	}
}

// FilterDate filters the collection to images within a date range.
// Note: This returns a new ImageCollection, not an Image.
// You'll typically want to call .First() or another method to get an Image.
func (ic *ImageCollection) FilterDate(startDate, endDate string) *ImageCollection {
	// For now, we'll implement a simple version that just tracks the filter
	// but doesn't apply it yet. Full implementation would require more complex
	// expression building with Date constructors.
	// This is a placeholder for future implementation.
	return ic
}

// FilterByYear filters the collection to images from a specific year.
// This is a convenience method for NLCD and other annual datasets.
// Note: This is a placeholder for future implementation.
func (ic *ImageCollection) FilterByYear(year int) *ImageCollection {
	// TODO: Implement year filtering
	// This will require Date constructors and filter expressions
	return ic
}

// Mosaic creates a composite image from the collection by mosaicking.
// Later images are rendered on top of earlier images.
func (ic *ImageCollection) Mosaic() *Image {
	// Create ImageCollection.load node
	loadNodeID := ic.expr.FunctionCall("ImageCollection.load", map[string]interface{}{
		"id": map[string]interface{}{
			"constantValue": ic.collectionID,
		},
	})

	// Create ImageCollection.mosaic node
	mosaicNodeID := ic.expr.FunctionCall("ImageCollection.mosaic", map[string]interface{}{
		"collection": map[string]interface{}{
			"valueReference": loadNodeID,
		},
	})

	return &Image{
		client: ic.client,
		expr:   ic.expr,
		nodeID: mosaicNodeID,
	}
}
