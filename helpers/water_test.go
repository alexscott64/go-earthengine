package helpers

import (
	"testing"
)

func TestWaterWithScaleOption(t *testing.T) {
	cfg := &waterConfig{scale: 30}
	WaterWithScale(100)(cfg)

	if cfg.scale != 100 {
		t.Errorf("scale = %f, want 100", cfg.scale)
	}
}

func TestWaterDetectionRequiresValidCoordinates(t *testing.T) {
	_, err := WaterDetection(nil, 100, -122)
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestWaterOccurrenceRequiresValidCoordinates(t *testing.T) {
	_, err := WaterOccurrence(nil, -95, 200)
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestWaterSeasonalityRequiresValidCoordinates(t *testing.T) {
	_, err := WaterSeasonality(nil, 45.5, 200)
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestWaterChangeRequiresValidCoordinates(t *testing.T) {
	_, err := WaterChange(nil, 91, -122)
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNewWaterDetectionQuery(t *testing.T) {
	query := NewWaterDetectionQuery(45.5152, -122.6784, WaterWithScale(50))

	if query == nil {
		t.Fatal("NewWaterDetectionQuery returned nil")
	}

	wq, ok := query.(*WaterDetectionQuery)
	if !ok {
		t.Fatal("Query is not a WaterDetectionQuery")
	}

	if wq.lat != 45.5152 {
		t.Errorf("lat = %f, want 45.5152", wq.lat)
	}
	if wq.lon != -122.6784 {
		t.Errorf("lon = %f, want -122.6784", wq.lon)
	}
}

func TestNewWaterOccurrenceQuery(t *testing.T) {
	query := NewWaterOccurrenceQuery(45.5152, -122.6784)

	if query == nil {
		t.Fatal("NewWaterOccurrenceQuery returned nil")
	}

	wq, ok := query.(*WaterOccurrenceQuery)
	if !ok {
		t.Fatal("Query is not a WaterOccurrenceQuery")
	}

	if wq.lat != 45.5152 {
		t.Errorf("lat = %f, want 45.5152", wq.lat)
	}
}

func TestWaterDatasetConstants(t *testing.T) {
	if jrcWaterDatasetID == "" {
		t.Error("jrcWaterDatasetID is empty")
	}
	if jrcMonthlyWaterID == "" {
		t.Error("jrcMonthlyWaterID is empty")
	}
}
