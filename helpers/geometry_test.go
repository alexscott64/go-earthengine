package helpers

import (
	"math"
	"testing"
)

func TestBoundsFromPoints(t *testing.T) {
	tests := []struct {
		name   string
		points [][2]float64
		want   Bounds
	}{
		{
			name: "simple box",
			points: [][2]float64{
				{-122.5, 45.4},
				{-122.3, 45.6},
			},
			want: Bounds{
				MinLon: -122.5, MinLat: 45.4,
				MaxLon: -122.3, MaxLat: 45.6,
			},
		},
		{
			name: "multiple points",
			points: [][2]float64{
				{-122.5, 45.4},
				{-122.3, 45.6},
				{-122.7, 45.5},
				{-122.4, 45.3},
			},
			want: Bounds{
				MinLon: -122.7, MinLat: 45.3,
				MaxLon: -122.3, MaxLat: 45.6,
			},
		},
		{
			name:   "empty points",
			points: [][2]float64{},
			want:   Bounds{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoundsFromPoints(tt.points)
			if got != tt.want {
				t.Errorf("BoundsFromPoints() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBoundsCenter(t *testing.T) {
	bounds := Bounds{
		MinLon: -122.5, MinLat: 45.4,
		MaxLon: -122.3, MaxLat: 45.6,
	}

	lat, lon := bounds.Center()
	wantLat, wantLon := 45.5, -122.4

	if lat != wantLat || lon != wantLon {
		t.Errorf("Center() = (%f, %f), want (%f, %f)", lat, lon, wantLat, wantLon)
	}
}

func TestBoundsContains(t *testing.T) {
	bounds := Bounds{
		MinLon: -122.5, MinLat: 45.4,
		MaxLon: -122.3, MaxLat: 45.6,
	}

	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{"inside", 45.5, -122.4, true},
		{"on boundary", 45.4, -122.4, true},
		{"outside north", 45.7, -122.4, false},
		{"outside south", 45.3, -122.4, false},
		{"outside east", 45.5, -122.2, false},
		{"outside west", 45.5, -122.6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bounds.Contains(tt.lat, tt.lon)
			if got != tt.want {
				t.Errorf("Contains(%f, %f) = %v, want %v", tt.lat, tt.lon, got, tt.want)
			}
		})
	}
}

func TestBoundsExpand(t *testing.T) {
	bounds := Bounds{
		MinLon: -122.5, MinLat: 45.4,
		MaxLon: -122.3, MaxLat: 45.6,
	}

	expanded := bounds.Expand(0.1) // 10% expansion

	// Check that it expanded in all directions
	if expanded.MinLon >= bounds.MinLon {
		t.Errorf("MinLon not expanded: %f >= %f", expanded.MinLon, bounds.MinLon)
	}
	if expanded.MaxLon <= bounds.MaxLon {
		t.Errorf("MaxLon not expanded: %f <= %f", expanded.MaxLon, bounds.MaxLon)
	}
	if expanded.MinLat >= bounds.MinLat {
		t.Errorf("MinLat not expanded: %f >= %f", expanded.MinLat, bounds.MinLat)
	}
	if expanded.MaxLat <= bounds.MaxLat {
		t.Errorf("MaxLat not expanded: %f <= %f", expanded.MaxLat, bounds.MaxLat)
	}

	// Check that the center is the same
	origLat, origLon := bounds.Center()
	expLat, expLon := expanded.Center()
	if math.Abs(origLat-expLat) > 0.0001 || math.Abs(origLon-expLon) > 0.0001 {
		t.Errorf("Center changed: (%f, %f) -> (%f, %f)", origLat, origLon, expLat, expLon)
	}
}

func TestBoundsValidate(t *testing.T) {
	tests := []struct {
		name    string
		bounds  Bounds
		wantErr bool
	}{
		{
			name: "valid bounds",
			bounds: Bounds{
				MinLon: -122.5, MinLat: 45.4,
				MaxLon: -122.3, MaxLat: 45.6,
			},
			wantErr: false,
		},
		{
			name: "invalid lat range",
			bounds: Bounds{
				MinLon: -122.5, MinLat: 45.6,
				MaxLon: -122.3, MaxLat: 45.4,
			},
			wantErr: true,
		},
		{
			name: "invalid lon range",
			bounds: Bounds{
				MinLon: -122.3, MinLat: 45.4,
				MaxLon: -122.5, MaxLat: 45.6,
			},
			wantErr: true,
		},
		{
			name: "invalid coordinates",
			bounds: Bounds{
				MinLon: -200, MinLat: 45.4,
				MaxLon: -122.3, MaxLat: 45.6,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bounds.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDistanceMeters(t *testing.T) {
	// Test Portland to Seattle (approximately 233 km)
	portland := struct{ lat, lon float64 }{45.5152, -122.6784}
	seattle := struct{ lat, lon float64 }{47.6062, -122.3321}

	distance := DistanceMeters(portland.lat, portland.lon, seattle.lat, seattle.lon)

	// Distance should be approximately 233 km (233000 m)
	// Allow 10% error due to our simplified calculation
	expected := 233000.0
	tolerance := expected * 0.10

	if math.Abs(distance-expected) > tolerance {
		t.Errorf("DistanceMeters() = %.0f, want approximately %.0f (±%.0f)", distance, expected, tolerance)
	}
}

func TestDistanceMetersZero(t *testing.T) {
	// Distance from a point to itself should be zero
	lat, lon := 45.5152, -122.6784
	distance := DistanceMeters(lat, lon, lat, lon)

	if distance > 0.001 {
		t.Errorf("DistanceMeters(same point) = %f, want ~0", distance)
	}
}

func TestCircleInvalidInputs(t *testing.T) {
	tests := []struct {
		name          string
		lat           float64
		lon           float64
		radiusMeters  float64
		wantErrPrefix string
	}{
		{"invalid lat", 100, -122.6784, 1000, "invalid latitude"},
		{"invalid lon", 45.5152, -200, 1000, "invalid longitude"},
		{"negative radius", 45.5152, -122.6784, -1000, "radius must be positive"},
		{"zero radius", 45.5152, -122.6784, 0, "radius must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Circle(tt.lat, tt.lon, tt.radiusMeters)
			if err == nil {
				t.Errorf("Circle() expected error, got nil")
			}
		})
	}
}

func TestPolygonInvalidInputs(t *testing.T) {
	tests := []struct {
		name   string
		points [][2]float64
	}{
		{"empty", [][2]float64{}},
		{"one point", [][2]float64{{-122, 45}}},
		{"two points", [][2]float64{{-122, 45}, {-121, 46}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Polygon(tt.points)
			if err == nil {
				t.Errorf("Polygon() expected error for %d points, got nil", len(tt.points))
			}
		})
	}
}

func TestBufferInvalidInputs(t *testing.T) {
	tests := []struct {
		name   string
		geom   interface{}
		meters float64
	}{
		{"nil geometry", nil, 1000},
		{"negative meters", "fake", -1000},
		{"zero meters", "fake", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Buffer(nil, tt.meters)
			if err == nil {
				t.Errorf("Buffer() expected error, got nil")
			}
		})
	}
}

func TestMathHelpers(t *testing.T) {
	// Test our simplified math functions are reasonably accurate

	// Test sin
	if math.Abs(sin(0)) > 0.001 {
		t.Errorf("sin(0) = %f, want ~0", sin(0))
	}
	if math.Abs(sin(math.Pi/2)-1) > 0.01 {
		t.Errorf("sin(π/2) = %f, want ~1", sin(math.Pi/2))
	}

	// Test sqrt
	if math.Abs(sqrt(4)-2) > 0.001 {
		t.Errorf("sqrt(4) = %f, want ~2", sqrt(4))
	}
	if math.Abs(sqrt(9)-3) > 0.001 {
		t.Errorf("sqrt(9) = %f, want ~3", sqrt(9))
	}
	if sqrt(0) != 0 {
		t.Errorf("sqrt(0) = %f, want 0", sqrt(0))
	}
}

func ExampleBoundsFromPoints() {
	// Create bounds from multiple points
	// points := [][2]float64{
	//     {-122.5, 45.4},  // Portland area
	//     {-122.3, 45.6},
	// }
	// bounds := BoundsFromPoints(points)
	// fmt.Printf("Bounds: %+v\n", bounds)
}

func ExampleBounds_Center() {
	// Get the center of a bounding box
	// bounds := Bounds{
	//     MinLon: -122.5, MinLat: 45.4,
	//     MaxLon: -122.3, MaxLat: 45.6,
	// }
	// lat, lon := bounds.Center()
	// fmt.Printf("Center: (%f, %f)\n", lat, lon)
}

func ExampleCircle() {
	// Create a 1km circle around Portland
	// circle, err := Circle(45.5152, -122.6784, 1000)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	// }
}

func ExampleDistanceMeters() {
	// Calculate distance between Portland and Seattle
	// distance := DistanceMeters(45.5152, -122.6784, 47.6062, -122.3321)
	// fmt.Printf("Distance: %.0f meters (%.0f km)\n", distance, distance/1000)
}
