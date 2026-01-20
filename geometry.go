package earthengine

// Geometry represents a geometric object in Earth Engine.
type Geometry interface {
	// NodeID returns the node ID for this geometry in the expression graph.
	NodeID(expr *ExpressionBuilder) string
}

// Point represents a geographic point.
type Point struct {
	Longitude float64
	Latitude  float64
}

// NewPoint creates a new Point geometry.
func NewPoint(longitude, latitude float64) Point {
	return Point{
		Longitude: longitude,
		Latitude:  latitude,
	}
}

// NodeID implements the Geometry interface for Point.
func (p Point) NodeID(expr *ExpressionBuilder) string {
	// Create coordinates array [longitude, latitude]
	coordinates := []interface{}{p.Longitude, p.Latitude}

	// Create GeometryConstructors.Point node
	return expr.FunctionCall(AlgorithmGeometryPoint, map[string]interface{}{
		"coordinates": map[string]interface{}{
			"constantValue": coordinates,
		},
	})
}
