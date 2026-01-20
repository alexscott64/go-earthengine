package earthengine

import (
	"context"
	"fmt"
)

// Image represents an Earth Engine Image with chainable operations.
type Image struct {
	client *Client
	expr   *ExpressionBuilder
	nodeID string // Node ID representing this image in the expression graph
}

// NewImage creates a new Image by loading it from the Earth Engine catalog.
func (c *Client) Image(imageID string) *Image {
	expr := NewExpressionBuilder()

	// Create Image.load node
	loadNodeID := expr.FunctionCall(AlgorithmImageLoad, map[string]interface{}{
		"id": map[string]interface{}{
			"constantValue": imageID,
		},
	})

	return &Image{
		client: c,
		expr:   expr,
		nodeID: loadNodeID,
	}
}

// Select selects specific bands from the image.
func (img *Image) Select(bands ...string) *Image {
	// Create band selectors array
	bandSelectors := make([]interface{}, len(bands))
	for i, band := range bands {
		bandSelectors[i] = band
	}

	// Create Image.select node
	selectNodeID := img.expr.FunctionCall(AlgorithmImageSelect, map[string]interface{}{
		"input": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"bandSelectors": map[string]interface{}{
			"constantValue": bandSelectors,
		},
	})

	// Return new Image with updated node ID
	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: selectNodeID,
	}
}

// ReduceRegionOperation represents a reduce region operation on an image.
type ReduceRegionOperation struct {
	image    *Image
	geometry string // Node ID for geometry
	reducer  string // Node ID for reducer
	scale    *float64
}

// ReduceRegion starts a reduce region operation.
func (img *Image) ReduceRegion(geom Geometry, reducer Reducer, opts ...ReduceRegionOption) *ReduceRegionOperation {
	op := &ReduceRegionOperation{
		image:    img,
		geometry: geom.NodeID(img.expr),
		reducer:  reducer.NodeID(img.expr),
	}

	// Apply options
	for _, opt := range opts {
		opt(op)
	}

	return op
}

// ReduceRegionOption is a function that configures a ReduceRegionOperation.
type ReduceRegionOption func(*ReduceRegionOperation)

// Scale sets the scale (resolution in meters) for the reduce region operation.
func Scale(meters float64) ReduceRegionOption {
	return func(op *ReduceRegionOperation) {
		op.scale = &meters
	}
}

// Compute executes the reduce region operation and returns the result.
func (op *ReduceRegionOperation) Compute(ctx context.Context) (map[string]interface{}, error) {
	// Build the reduceRegion function call arguments
	args := map[string]interface{}{
		"image": map[string]interface{}{
			"valueReference": op.image.nodeID,
		},
		"geometry": map[string]interface{}{
			"valueReference": op.geometry,
		},
		"reducer": map[string]interface{}{
			"valueReference": op.reducer,
		},
	}

	// Add optional scale parameter
	if op.scale != nil {
		args["scale"] = map[string]interface{}{
			"constantValue": *op.scale,
		}
	}

	// Create reduceRegion node
	reduceNodeID := op.image.expr.FunctionCall(AlgorithmImageReduceRegion, args)

	// Build the expression
	expr := op.image.expr.Build(reduceNodeID)

	// Execute the expression
	result, err := op.image.client.ComputeValue(ctx, expr)
	if err != nil {
		return nil, err
	}

	// Convert result to map
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return resultMap, nil
}

// ComputeFloat executes the reduce region operation and returns a single float64 value.
// This is useful when you know the result will be a single band value.
func (op *ReduceRegionOperation) ComputeFloat(ctx context.Context) (float64, error) {
	result, err := op.Compute(ctx)
	if err != nil {
		return 0, err
	}

	// Try to extract a single numeric value
	// The result structure depends on the reducer and bands
	for _, v := range result {
		if num, ok := v.(float64); ok {
			return num, nil
		}
	}

	return 0, fmt.Errorf("no numeric value found in result: %v", result)
}

// Add performs element-wise addition with another image.
func (img *Image) Add(other *Image) *Image {
	addNodeID := img.expr.FunctionCall(AlgorithmImageAdd, map[string]interface{}{
		"image1": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"image2": map[string]interface{}{
			"valueReference": other.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: addNodeID,
	}
}

// Subtract performs element-wise subtraction with another image.
func (img *Image) Subtract(other *Image) *Image {
	subtractNodeID := img.expr.FunctionCall(AlgorithmImageSubtract, map[string]interface{}{
		"image1": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"image2": map[string]interface{}{
			"valueReference": other.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: subtractNodeID,
	}
}

// Multiply performs element-wise multiplication with another image.
func (img *Image) Multiply(other *Image) *Image {
	multiplyNodeID := img.expr.FunctionCall(AlgorithmImageMultiply, map[string]interface{}{
		"image1": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"image2": map[string]interface{}{
			"valueReference": other.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: multiplyNodeID,
	}
}

// Divide performs element-wise division with another image.
func (img *Image) Divide(other *Image) *Image {
	divideNodeID := img.expr.FunctionCall(AlgorithmImageDivide, map[string]interface{}{
		"image1": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"image2": map[string]interface{}{
			"valueReference": other.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: divideNodeID,
	}
}

// NormalizedDifference computes the normalized difference between two bands: (b1 - b2) / (b1 + b2).
// This is commonly used for vegetation indices (NDVI), water indices (NDWI), etc.
//
// Example:
//
//	// Calculate NDVI from Sentinel-2
//	ndvi := image.Select("B8", "B4").NormalizedDifference()
func (img *Image) NormalizedDifference() *Image {
	ndNodeID := img.expr.FunctionCall(AlgorithmImageNormalizedDiff, map[string]interface{}{
		"input": map[string]interface{}{
			"valueReference": img.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: ndNodeID,
	}
}

// Expression evaluates a mathematical expression on an image.
//
// Example:
//
//	// Calculate EVI: 2.5 * ((NIR - RED) / (NIR + 6*RED - 7.5*BLUE + 1))
//	evi := image.Expression("2.5 * ((NIR - RED) / (NIR + 6*RED - 7.5*BLUE + 1))", map[string]interface{}{
//	    "NIR":  image.Select("B8"),
//	    "RED":  image.Select("B4"),
//	    "BLUE": image.Select("B2"),
//	})
func (img *Image) Expression(expression string, vars map[string]interface{}) *Image {
	// Convert vars to expression references
	varRefs := make(map[string]interface{})
	for k, v := range vars {
		if imgVar, ok := v.(*Image); ok {
			varRefs[k] = map[string]interface{}{
				"valueReference": imgVar.nodeID,
			}
		} else {
			varRefs[k] = map[string]interface{}{
				"constantValue": v,
			}
		}
	}

	exprNodeID := img.expr.FunctionCall(AlgorithmImageExpression, map[string]interface{}{
		"image": map[string]interface{}{
			"valueReference": img.nodeID,
		},
		"expression": map[string]interface{}{
			"constantValue": expression,
		},
		"map": varRefs,
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: exprNodeID,
	}
}

// Terrain applies a terrain algorithm to an elevation image.
// Use AlgorithmTerrainSlope or AlgorithmTerrainAspect as the algorithm parameter.
func (img *Image) Terrain(algorithm string) *Image {
	terrainNodeID := img.expr.FunctionCall(algorithm, map[string]interface{}{
		"input": map[string]interface{}{
			"valueReference": img.nodeID,
		},
	})

	return &Image{
		client: img.client,
		expr:   img.expr,
		nodeID: terrainNodeID,
	}
}
