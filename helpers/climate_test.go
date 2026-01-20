package helpers

import (
	"testing"
)

func TestTerraClimateOption(t *testing.T) {
	opts := &ClimateOptions{}
	TerraClimate()(opts)

	if opts.dataset != terraClimateDatasetID {
		t.Errorf("dataset = %s, want %s", opts.dataset, terraClimateDatasetID)
	}
	if opts.scale != 4000 {
		t.Errorf("scale = %f, want 4000", opts.scale)
	}
}

func TestCHIRPSOption(t *testing.T) {
	opts := &ClimateOptions{}
	CHIRPS()(opts)

	if opts.dataset != chirpsDatasetID {
		t.Errorf("dataset = %s, want %s", opts.dataset, chirpsDatasetID)
	}
	if opts.scale != 5000 {
		t.Errorf("scale = %f, want 5000", opts.scale)
	}
}

func TestSMAPOption(t *testing.T) {
	opts := &ClimateOptions{}
	SMAP()(opts)

	if opts.dataset != smapDatasetID {
		t.Errorf("dataset = %s, want %s", opts.dataset, smapDatasetID)
	}
	if opts.scale != 9000 {
		t.Errorf("scale = %f, want 9000", opts.scale)
	}
}

func TestClimateDateRangeOption(t *testing.T) {
	opts := &ClimateOptions{}
	ClimateDateRange("2023-01-01", "2023-12-31")(opts)

	if opts.startDate != "2023-01-01" {
		t.Errorf("startDate = %s, want 2023-01-01", opts.startDate)
	}
	if opts.endDate != "2023-12-31" {
		t.Errorf("endDate = %s, want 2023-12-31", opts.endDate)
	}
}

func TestTemperatureRequiresDateRange(t *testing.T) {
	_, err := Temperature(nil, 45.5152, -122.6784)
	if err == nil {
		t.Error("Expected error when date range is missing")
	}
}

func TestPrecipitationRequiresDateRange(t *testing.T) {
	_, err := Precipitation(nil, 45.5152, -122.6784)
	if err == nil {
		t.Error("Expected error when date range is missing")
	}
}

func TestSoilMoistureRequiresDateRange(t *testing.T) {
	_, err := SoilMoisture(nil, 45.5152, -122.6784)
	if err == nil {
		t.Error("Expected error when date range is missing")
	}
}

func TestTemperatureInvalidCoordinates(t *testing.T) {
	_, err := Temperature(nil, 100, -122,
		ClimateDateRange("2023-01-01", "2023-12-31"))
	if err == nil {
		t.Error("Expected error for invalid coordinates")
	}
}

func TestNewTemperatureQuery(t *testing.T) {
	query := NewTemperatureQuery(45.5152, -122.6784,
		ClimateDateRange("2023-01-01", "2023-12-31"))

	if query == nil {
		t.Fatal("NewTemperatureQuery returned nil")
	}

	cq, ok := query.(*ClimateQuery)
	if !ok {
		t.Fatal("Query is not a ClimateQuery")
	}

	if cq.lat != 45.5152 {
		t.Errorf("lat = %f, want 45.5152", cq.lat)
	}
}

func TestNewPrecipitationQuery(t *testing.T) {
	query := NewPrecipitationQuery(45.5152, -122.6784,
		CHIRPS(),
		ClimateDateRange("2023-06-01", "2023-06-30"))

	if query == nil {
		t.Fatal("NewPrecipitationQuery returned nil")
	}
}

func TestNewSoilMoistureQuery(t *testing.T) {
	query := NewSoilMoistureQuery(45.5152, -122.6784,
		SMAP(),
		ClimateDateRange("2023-07-01", "2023-07-31"))

	if query == nil {
		t.Fatal("NewSoilMoistureQuery returned nil")
	}
}

func TestClimateDatasetConstants(t *testing.T) {
	if terraClimateDatasetID == "" {
		t.Error("terraClimateDatasetID is empty")
	}
	if chirpsDatasetID == "" {
		t.Error("chirpsDatasetID is empty")
	}
	if smapDatasetID == "" {
		t.Error("smapDatasetID is empty")
	}
}
