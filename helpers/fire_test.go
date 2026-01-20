package helpers

import (
	"testing"
)

func TestVIIRSOption(t *testing.T) {
	cfg := &fireConfig{}
	VIIRS()(cfg)

	if cfg.dataset != viirsFireDatasetID {
		t.Errorf("dataset = %s, want %s", cfg.dataset, viirsFireDatasetID)
	}
	if cfg.scale != 375 {
		t.Errorf("scale = %f, want 375", cfg.scale)
	}
}

func TestMODISFireOption(t *testing.T) {
	cfg := &fireConfig{}
	MODISFire()(cfg)

	if cfg.dataset != modisFireDatasetID {
		t.Errorf("dataset = %s, want %s", cfg.dataset, modisFireDatasetID)
	}
	if cfg.scale != 1000 {
		t.Errorf("scale = %f, want 1000", cfg.scale)
	}
}

func TestFireDateRangeOption(t *testing.T) {
	cfg := &fireConfig{}
	FireDateRange("2023-08-01", "2023-08-31")(cfg)

	if cfg.dateRange == nil {
		t.Fatal("dateRange is nil")
	}
	if cfg.dateRange.Start != "2023-08-01" {
		t.Errorf("Start = %s, want 2023-08-01", cfg.dateRange.Start)
	}
	if cfg.dateRange.End != "2023-08-31" {
		t.Errorf("End = %s, want 2023-08-31", cfg.dateRange.End)
	}
}

func TestActiveFireRequiresDateRange(t *testing.T) {
	_, err := ActiveFire(nil, 45.5152, -122.6784)
	if err == nil {
		t.Error("Expected error when date range is missing")
	}
}

func TestFireCountRequiresDateRange(t *testing.T) {
	_, err := FireCount(nil, 45.5152, -122.6784)
	if err == nil {
		t.Error("Expected error when date range is missing")
	}
}

func TestActiveFireRequiresValidCoordinates(t *testing.T) {
	_, err := ActiveFire(nil, 100, -122,
		FireDateRange("2023-08-01", "2023-08-31"))
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestBurnSeverityRequiresValidCoordinates(t *testing.T) {
	_, err := BurnSeverity(nil, -95, 200, "2023-08-15")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestDeltaNBRRequiresValidCoordinates(t *testing.T) {
	_, err := DeltaNBR(nil, 91, -122, "2023-07-01", "2023-09-01")
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNewActiveFireQuery(t *testing.T) {
	query := NewActiveFireQuery(45.5152, -122.6784,
		FireDateRange("2023-08-01", "2023-08-31"))

	if query == nil {
		t.Fatal("NewActiveFireQuery returned nil")
	}

	fq, ok := query.(*ActiveFireQuery)
	if !ok {
		t.Fatal("Query is not an ActiveFireQuery")
	}

	if fq.lat != 45.5152 {
		t.Errorf("lat = %f, want 45.5152", fq.lat)
	}
}

func TestNewBurnSeverityQuery(t *testing.T) {
	query := NewBurnSeverityQuery(45.5152, -122.6784, "2023-08-15")

	if query == nil {
		t.Fatal("NewBurnSeverityQuery returned nil")
	}

	bq, ok := query.(*BurnSeverityQuery)
	if !ok {
		t.Fatal("Query is not a BurnSeverityQuery")
	}

	if bq.lat != 45.5152 {
		t.Errorf("lat = %f, want 45.5152", bq.lat)
	}
	if bq.date != "2023-08-15" {
		t.Errorf("date = %s, want 2023-08-15", bq.date)
	}
}

func TestFireDatasetConstants(t *testing.T) {
	if viirsFireDatasetID == "" {
		t.Error("viirsFireDatasetID is empty")
	}
	if modisFireDatasetID == "" {
		t.Error("modisFireDatasetID is empty")
	}
	if landsat8SRID == "" {
		t.Error("landsat8SRID is empty")
	}
}
