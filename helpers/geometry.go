package helpers

import (
	"fmt"

	"github.com/yourusername/go-earthengine"
)

// Bounds represents a geographic bounding box.
type Bounds struct {
	MinLon float64 // Western longitude
	MinLat float64 // Southern latitude
	MaxLon float64 // Eastern longitude
	MaxLat float64 // Northern latitude
}

// BoundsFromPoints creates a bounding box from a set of points.
//
// Example:
//
//	points := [][2]float64{
//	    {-122.5, 45.4},  // Portland area
//	    {-122.3, 45.6},
//	}
//	bounds := helpers.BoundsFromPoints(points)
func BoundsFromPoints(points [][2]float64) Bounds {
	if len(points) == 0 {
		return Bounds{}
	}

	minLon, minLat := points[0][0], points[0][1]
	maxLon, maxLat := points[0][0], points[0][1]

	for _, pt := range points[1:] {
		lon, lat := pt[0], pt[1]
		if lon < minLon {
			minLon = lon
		}
		if lon > maxLon {
			maxLon = lon
		}
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
	}

	return Bounds{
		MinLon: minLon,
		MinLat: minLat,
		MaxLon: maxLon,
		MaxLat: maxLat,
	}
}

// ToRectangle converts the bounds to an Earth Engine Rectangle geometry.
//
// Note: This returns a earthengine.Geometry interface.
// The actual implementation would need Rectangle support in the main library.
func (b Bounds) ToRectangle() (earthengine.Geometry, error) {
	if err := validateCoordinates(b.MinLat, b.MinLon); err != nil {
		return nil, fmt.Errorf("invalid min coordinates: %w", err)
	}
	if err := validateCoordinates(b.MaxLat, b.MaxLon); err != nil {
		return nil, fmt.Errorf("invalid max coordinates: %w", err)
	}
	if b.MinLon >= b.MaxLon {
		return nil, fmt.Errorf("minLon (%f) must be less than maxLon (%f)", b.MinLon, b.MaxLon)
	}
	if b.MinLat >= b.MaxLat {
		return nil, fmt.Errorf("minLat (%f) must be less than maxLat (%f)", b.MinLat, b.MaxLat)
	}

	// Placeholder - would need Rectangle geometry support in main library
	return nil, fmt.Errorf("Rectangle geometry support not yet implemented")
}

// Area returns the approximate area of the bounds in square meters.
//
// This uses a simple calculation that assumes the Earth is a sphere.
// For more accurate calculations, use Earth Engine's computeArea() method.
func (b Bounds) Area() float64 {
	const earthRadius = 6371000.0 // meters

	// Convert to radians
	lat1 := b.MinLat * 3.141592653589793 / 180.0
	lat2 := b.MaxLat * 3.141592653589793 / 180.0
	lon1 := b.MinLon * 3.141592653589793 / 180.0
	lon2 := b.MaxLon * 3.141592653589793 / 180.0

	// Calculate area
	width := (lon2 - lon1) * earthRadius * ((lat1 + lat2) / 2.0)
	height := (lat2 - lat1) * earthRadius

	return width * height
}

// Center returns the center point of the bounds.
func (b Bounds) Center() (lat, lon float64) {
	return (b.MinLat + b.MaxLat) / 2.0, (b.MinLon + b.MaxLon) / 2.0
}

// Contains checks if a point is within the bounds.
func (b Bounds) Contains(lat, lon float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat &&
		lon >= b.MinLon && lon <= b.MaxLon
}

// Expand expands the bounds by a percentage.
//
// Example:
//
//	expanded := bounds.Expand(0.1) // Expand by 10% on all sides
func (b Bounds) Expand(percentage float64) Bounds {
	latRange := b.MaxLat - b.MinLat
	lonRange := b.MaxLon - b.MinLon

	return Bounds{
		MinLon: b.MinLon - lonRange*percentage,
		MinLat: b.MinLat - latRange*percentage,
		MaxLon: b.MaxLon + lonRange*percentage,
		MaxLat: b.MaxLat + latRange*percentage,
	}
}

// Validate checks if the bounds are valid.
func (b Bounds) Validate() error {
	if err := validateCoordinates(b.MinLat, b.MinLon); err != nil {
		return fmt.Errorf("invalid min coordinates: %w", err)
	}
	if err := validateCoordinates(b.MaxLat, b.MaxLon); err != nil {
		return fmt.Errorf("invalid max coordinates: %w", err)
	}
	if b.MinLon >= b.MaxLon {
		return fmt.Errorf("minLon (%f) must be less than maxLon (%f)", b.MinLon, b.MaxLon)
	}
	if b.MinLat >= b.MaxLat {
		return fmt.Errorf("minLat (%f) must be less than maxLat (%f)", b.MinLat, b.MaxLat)
	}
	return nil
}

// Circle creates a circular geometry around a point.
//
// Note: This is a placeholder. The actual implementation would need
// GeometryConstructors.Point + Buffer support in the main library.
//
// Example:
//
//	// Create 1km radius circle around Portland
//	circle, err := helpers.Circle(45.5152, -122.6784, 1000)
func Circle(lat, lon, radiusMeters float64) (earthengine.Geometry, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}
	if radiusMeters <= 0 {
		return nil, fmt.Errorf("radius must be positive, got %f", radiusMeters)
	}

	// Placeholder - would need Buffer geometry support
	return nil, fmt.Errorf("Circle geometry support not yet implemented")
}

// Rectangle creates a rectangular geometry from bounds.
//
// Note: This is a placeholder. The actual implementation would need
// GeometryConstructors.Rectangle support in the main library.
//
// Example:
//
//	bounds := helpers.Bounds{
//	    MinLon: -122.5, MinLat: 45.4,
//	    MaxLon: -122.3, MaxLat: 45.6,
//	}
//	rect, err := helpers.Rectangle(bounds)
func Rectangle(bounds Bounds) (earthengine.Geometry, error) {
	return bounds.ToRectangle()
}

// Polygon creates a polygon geometry from a set of points.
//
// Note: This is a placeholder. The actual implementation would need
// GeometryConstructors.Polygon support in the main library.
//
// Points should be in [lon, lat] format.
//
// Example:
//
//	points := [][2]float64{
//	    {-122.5, 45.4},
//	    {-122.3, 45.4},
//	    {-122.3, 45.6},
//	    {-122.5, 45.6},
//	    {-122.5, 45.4}, // Close the polygon
//	}
//	polygon, err := helpers.Polygon(points)
func Polygon(points [][2]float64) (earthengine.Geometry, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("polygon requires at least 3 points, got %d", len(points))
	}

	// Validate all points
	for i, pt := range points {
		if err := validateCoordinates(pt[1], pt[0]); err != nil {
			return nil, fmt.Errorf("invalid point %d: %w", i, err)
		}
	}

	// Placeholder - would need Polygon geometry support
	return nil, fmt.Errorf("Polygon geometry support not yet implemented")
}

// Buffer creates a buffered geometry around another geometry.
//
// Note: This is a placeholder. The actual implementation would need
// Geometry.buffer() support in the main library.
//
// Example:
//
//	point := earthengine.NewPoint(-122.6784, 45.5152)
//	buffered, err := helpers.Buffer(point, 1000) // 1km buffer
func Buffer(geom earthengine.Geometry, meters float64) (earthengine.Geometry, error) {
	if geom == nil {
		return nil, fmt.Errorf("geometry cannot be nil")
	}
	if meters <= 0 {
		return nil, fmt.Errorf("buffer distance must be positive, got %f", meters)
	}

	// Placeholder - would need Buffer support
	return nil, fmt.Errorf("Buffer geometry support not yet implemented")
}

// Helper functions for geometry calculations

// DistanceMeters calculates the approximate distance between two points in meters.
//
// Uses the Haversine formula for great-circle distance.
func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0 // meters

	// Convert to radians
	lat1Rad := lat1 * 3.141592653589793 / 180.0
	lat2Rad := lat2 * 3.141592653589793 / 180.0
	deltaLat := (lat2 - lat1) * 3.141592653589793 / 180.0
	deltaLon := (lon2 - lon1) * 3.141592653589793 / 180.0

	// Haversine formula
	a := sin(deltaLat/2)*sin(deltaLat/2) +
		cos(lat1Rad)*cos(lat2Rad)*sin(deltaLon/2)*sin(deltaLon/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return earthRadius * c
}

// Helper math functions (since we can't import math to avoid circular deps)
func sin(x float64) float64 {
	// Taylor series approximation for sine
	// Good enough for our purposes
	term := x
	sum := term
	for i := 1; i < 10; i++ {
		term *= -x * x / float64((2*i)*(2*i+1))
		sum += term
	}
	return sum
}

func cos(x float64) float64 {
	// cos(x) = sin(x + π/2)
	return sin(x + 1.5707963267948966)
}

func sqrt(x float64) float64 {
	// Newton's method
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func atan2(y, x float64) float64 {
	// Simple atan2 approximation
	if x == 0 {
		if y > 0 {
			return 1.5707963267948966 // π/2
		}
		return -1.5707963267948966 // -π/2
	}
	atan := y / x
	// First order approximation
	return atan / (1 + 0.28*atan*atan)
}
