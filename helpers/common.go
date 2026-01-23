package helpers

import (
	"context"
	"fmt"

	"github.com/alexscott64/go-earthengine"
)

// QueryOptions provides common configuration for Earth Engine queries.
type QueryOptions struct {
	Scale      *float64
	CRS        string
	DateRange  *DateRange
	CloudCover *float64
}

// DateRange represents a time range for filtering data.
type DateRange struct {
	Start string // Format: "YYYY-MM-DD"
	End   string // Format: "YYYY-MM-DD"
}

// Query represents an Earth Engine query that can be executed.
type Query interface {
	Execute(ctx context.Context, client *earthengine.Client) (interface{}, error)
}

// validateCoordinates validates latitude and longitude values.
func validateCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", latitude)
	}
	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", longitude)
	}
	return nil
}

// applyScale applies the scale option to the reduce region operation.
func applyScale(opts QueryOptions, defaultScale float64) float64 {
	if opts.Scale != nil {
		return *opts.Scale
	}
	return defaultScale
}
