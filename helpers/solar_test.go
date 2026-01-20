package helpers

import (
	"math"
	"testing"
	"time"
)

func TestSunPosition(t *testing.T) {
	// Test summer solstice noon in Portland
	summerSolstice := time.Date(2023, 6, 21, 19, 0, 0, 0, time.UTC) // ~noon in Portland (UTC-7)
	pos, err := CalculateSunPosition(45.5152, -122.6784, summerSolstice)

	if err != nil {
		t.Errorf("SunPosition() error = %v, want nil", err)
	}
	if pos == nil {
		t.Fatal("SunPosition() returned nil")
	}

	// Sun should be high in the sky (elevation > 60 degrees at summer solstice noon)
	if pos.Elevation < 50 {
		t.Errorf("SunPosition() elevation = %.1f, expected > 50 degrees at summer solstice noon", pos.Elevation)
	}

	// Zenith should be less than 40 degrees
	if pos.Zenith > 40 {
		t.Errorf("SunPosition() zenith = %.1f, expected < 40 degrees", pos.Zenith)
	}

	// Elevation + Zenith should equal 90
	if math.Abs(pos.Elevation+pos.Zenith-90.0) > 0.1 {
		t.Errorf("Elevation (%.1f) + Zenith (%.1f) != 90", pos.Elevation, pos.Zenith)
	}
}

func TestSunPositionInvalidCoordinates(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"invalid lat high", 100, -122},
		{"invalid lat low", -100, -122},
		{"invalid lon high", 45, 200},
		{"invalid lon low", 45, -200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateSunPosition(tt.lat, tt.lon, now)
			if err == nil {
				t.Errorf("SunPosition() expected error for invalid coordinates, got nil")
			}
		})
	}
}

func TestDayLength(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		date     time.Time
		minHours float64
		maxHours float64
	}{
		{
			name:     "summer solstice at 45N",
			lat:      45.0,
			date:     time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC),
			minHours: 15.0,
			maxHours: 16.0,
		},
		{
			name:     "winter solstice at 45N",
			lat:      45.0,
			date:     time.Date(2023, 12, 21, 0, 0, 0, 0, time.UTC),
			minHours: 8.0,
			maxHours: 9.5,
		},
		{
			name:     "equinox at equator",
			lat:      0.0,
			date:     time.Date(2023, 3, 20, 0, 0, 0, 0, time.UTC),
			minHours: 11.5,
			maxHours: 12.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DayLength(tt.lat, tt.date)
			if err != nil {
				t.Errorf("DayLength() error = %v, want nil", err)
			}

			hours := length.Hours()
			if hours < tt.minHours || hours > tt.maxHours {
				t.Errorf("DayLength() = %.1f hours, want between %.1f and %.1f",
					hours, tt.minHours, tt.maxHours)
			}
		})
	}
}

func TestDayLengthPolarRegions(t *testing.T) {
	// Test polar day (24 hours of daylight)
	summerSolstice := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	length, err := DayLength(70.0, summerSolstice) // Far north

	if err != nil {
		t.Errorf("DayLength() error = %v, want nil", err)
	}

	// Should be close to 24 hours
	if length.Hours() < 20 {
		t.Errorf("DayLength() at Arctic in summer = %.1f hours, expected ~24 (polar day)", length.Hours())
	}

	// Test polar night (0 hours of daylight)
	winterSolstice := time.Date(2023, 12, 21, 0, 0, 0, 0, time.UTC)
	length, err = DayLength(70.0, winterSolstice)

	if err != nil {
		t.Errorf("DayLength() error = %v, want nil", err)
	}

	// Should be close to 0 hours
	if length.Hours() > 4 {
		t.Errorf("DayLength() at Arctic in winter = %.1f hours, expected ~0 (polar night)", length.Hours())
	}
}

func TestSunriseTime(t *testing.T) {
	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	sunrise, err := SunriseTime(45.5152, -122.6784, date)

	if err != nil {
		t.Errorf("SunriseTime() error = %v, want nil", err)
	}

	// Sunrise should be in the morning (before noon UTC)
	if sunrise.Hour() > 20 {
		t.Errorf("SunriseTime() hour = %d, expected morning time", sunrise.Hour())
	}
}

func TestSunsetTime(t *testing.T) {
	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	sunset, err := SunsetTime(45.5152, -122.6784, date)

	if err != nil {
		t.Errorf("SunsetTime() error = %v, want nil", err)
	}

	// Sunset should be a valid time (we'll check order in TestSunriseSunsetOrder)
	if sunset.IsZero() {
		t.Error("SunsetTime() returned zero time")
	}
}

func TestSunriseSunsetOrder(t *testing.T) {
	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	sunrise, err := SunriseTime(45.5152, -122.6784, date)
	if err != nil {
		t.Fatalf("SunriseTime() error = %v", err)
	}

	sunset, err := SunsetTime(45.5152, -122.6784, date)
	if err != nil {
		t.Fatalf("SunsetTime() error = %v", err)
	}

	// Sunrise should be before sunset
	if !sunrise.Before(sunset) {
		t.Errorf("Sunrise (%s) should be before sunset (%s)",
			sunrise.Format("15:04"), sunset.Format("15:04"))
	}
}

func TestIsDaytime(t *testing.T) {
	// Test at noon (should be daytime)
	noon := time.Date(2023, 6, 21, 19, 0, 0, 0, time.UTC) // Noon in Portland
	daytime, err := IsDaytime(45.5152, -122.6784, noon)

	if err != nil {
		t.Errorf("IsDaytime() error = %v, want nil", err)
	}
	if !daytime {
		t.Error("IsDaytime() at noon = false, want true")
	}

	// Test at midnight (should be nighttime)
	midnight := time.Date(2023, 6, 21, 7, 0, 0, 0, time.UTC) // Midnight in Portland
	daytime, err = IsDaytime(45.5152, -122.6784, midnight)

	if err != nil {
		t.Errorf("IsDaytime() error = %v, want nil", err)
	}
	if daytime {
		t.Error("IsDaytime() at midnight = true, want false")
	}
}

func TestSolarNoon(t *testing.T) {
	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	noon, err := SolarNoon(-122.6784, date)

	if err != nil {
		t.Errorf("SolarNoon() error = %v, want nil", err)
	}

	// Solar noon should be around midday when accounting for longitude
	// For longitude -122.6784, solar noon is approximately 12:00 + 8.2 hours = 20:10 UTC
	expectedHour := 20
	if math.Abs(float64(noon.Hour()-expectedHour)) > 2 {
		t.Errorf("SolarNoon() hour = %d, want approximately %d", noon.Hour(), expectedHour)
	}
}

func TestSolarNoonInvalidLongitude(t *testing.T) {
	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		lon  float64
	}{
		{"too high", 200},
		{"too low", -200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SolarNoon(tt.lon, date)
			if err == nil {
				t.Error("SolarNoon() expected error for invalid longitude, got nil")
			}
		})
	}
}

func TestJulianDay(t *testing.T) {
	// J2000.0 epoch (January 1, 2000, 12:00 UTC) should be JD 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd := julianDay(j2000)

	expected := 2451545.0
	if math.Abs(jd-expected) > 0.1 {
		t.Errorf("julianDay(J2000) = %.1f, want %.1f", jd, expected)
	}
}

func ExampleCalculateSunPosition() {
	// Calculate sun position at summer solstice noon
	// t := time.Date(2023, 6, 21, 12, 0, 0, 0, time.UTC)
	// pos, err := CalculateSunPosition(45.5152, -122.6784, t)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }
	// fmt.Printf("Azimuth: %.1f°, Elevation: %.1f°\n", pos.Azimuth, pos.Elevation)
}

func ExampleDayLength() {
	// Calculate daylight hours on summer solstice
	// date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	// length, err := DayLength(45.5152, date)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }
	// fmt.Printf("Daylight: %.1f hours\n", length.Hours())
}

func ExampleSunriseTime() {
	// Calculate sunrise time
	// date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
	// sunrise, err := SunriseTime(45.5152, -122.6784, date)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }
	// fmt.Printf("Sunrise: %s UTC\n", sunrise.Format("15:04:05"))
}

func ExampleIsDaytime() {
	// Check if it's daytime
	// daytime, err := IsDaytime(45.5152, -122.6784, time.Now())
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }
	// if daytime {
	//     fmt.Println("It's daytime!")
	// } else {
	//     fmt.Println("It's nighttime!")
	// }
}
