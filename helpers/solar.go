package helpers

import (
	"fmt"
	"math"
	"time"
)

// SunPosition represents the position of the sun in the sky.
type SunPosition struct {
	Azimuth   float64 // Degrees from north (0-360)
	Elevation float64 // Degrees above horizon (-90 to 90)
	Zenith    float64 // Degrees from zenith (0-180)
}

// CalculateSunPosition calculates the sun's position at a given location and time.
//
// Uses a simplified algorithm suitable for most applications.
// For high-precision requirements, consider using a dedicated astronomy library.
//
// Returns:
//   - Azimuth: compass direction (0=North, 90=East, 180=South, 270=West)
//   - Elevation: angle above horizon (negative = below horizon)
//   - Zenith: angle from directly overhead
//
// Example:
//
//	t := time.Date(2023, 6, 21, 12, 0, 0, 0, time.UTC) // Summer solstice noon
//	pos, err := helpers.CalculateSunPosition(45.5152, -122.6784, t)
//	fmt.Printf("Azimuth: %.1f°, Elevation: %.1f°\n", pos.Azimuth, pos.Elevation)
func CalculateSunPosition(lat, lon float64, t time.Time) (*SunPosition, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	// Convert to radians
	latRad := lat * math.Pi / 180.0

	// Calculate Julian day
	jd := julianDay(t)

	// Calculate solar declination
	n := jd - 2451545.0 // Days since J2000.0
	L := math.Mod(280.460+0.9856474*n, 360.0)
	g := math.Mod(357.528+0.9856003*n, 360.0) * math.Pi / 180.0

	lambda := (L + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)) * math.Pi / 180.0
	epsilon := 23.439 * math.Pi / 180.0
	declination := math.Asin(math.Sin(epsilon) * math.Sin(lambda))

	// Calculate hour angle
	utcHours := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0
	hourAngle := (15.0*(utcHours-12.0) + lon) * math.Pi / 180.0

	// Calculate elevation
	sinElevation := math.Sin(latRad)*math.Sin(declination) +
		math.Cos(latRad)*math.Cos(declination)*math.Cos(hourAngle)
	elevation := math.Asin(sinElevation) * 180.0 / math.Pi

	// Calculate azimuth
	cosAzimuth := (math.Sin(declination) - math.Sin(latRad)*sinElevation) /
		(math.Cos(latRad) * math.Cos(math.Asin(sinElevation)))

	// Clamp to avoid numerical issues
	if cosAzimuth > 1.0 {
		cosAzimuth = 1.0
	}
	if cosAzimuth < -1.0 {
		cosAzimuth = -1.0
	}

	azimuth := math.Acos(cosAzimuth) * 180.0 / math.Pi

	// Adjust azimuth based on hour angle
	if hourAngle > 0 {
		azimuth = 360.0 - azimuth
	}

	return &SunPosition{
		Azimuth:   azimuth,
		Elevation: elevation,
		Zenith:    90.0 - elevation,
	}, nil
}

// DayLength calculates the length of daylight at a location on a given date.
//
// Returns the duration of daylight (sunrise to sunset).
//
// Note: This uses a simplified calculation. Atmospheric refraction and
// the finite size of the sun's disc can affect actual sunrise/sunset times.
//
// Example:
//
//	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC) // Summer solstice
//	length, err := helpers.DayLength(45.5152, date)
//	fmt.Printf("Daylight: %.1f hours\n", length.Hours())
func DayLength(lat float64, date time.Time) (time.Duration, error) {
	if lat < -90 || lat > 90 {
		return 0, fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", lat)
	}

	// Calculate solar declination for this date
	jd := julianDay(date)
	n := jd - 2451545.0
	L := math.Mod(280.460+0.9856474*n, 360.0)
	g := math.Mod(357.528+0.9856003*n, 360.0) * math.Pi / 180.0

	lambda := (L + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)) * math.Pi / 180.0
	epsilon := 23.439 * math.Pi / 180.0
	declination := math.Asin(math.Sin(epsilon) * math.Sin(lambda))

	// Calculate hour angle at sunset
	latRad := lat * math.Pi / 180.0
	cosHourAngle := -math.Tan(latRad) * math.Tan(declination)

	// Check for polar day/night
	if cosHourAngle > 1.0 {
		// Polar night - no sunrise
		return 0, nil
	}
	if cosHourAngle < -1.0 {
		// Polar day - 24 hours of daylight
		return 24 * time.Hour, nil
	}

	hourAngle := math.Acos(cosHourAngle)
	daylightHours := 2.0 * hourAngle * 12.0 / math.Pi

	return time.Duration(daylightHours * float64(time.Hour)), nil
}

// SunriseTime calculates the sunrise time at a location on a given date.
//
// Returns the time of sunrise in UTC.
//
// Example:
//
//	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
//	sunrise, err := helpers.SunriseTime(45.5152, -122.6784, date)
//	fmt.Printf("Sunrise: %s UTC\n", sunrise.Format("15:04:05"))
func SunriseTime(lat, lon float64, date time.Time) (time.Time, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return time.Time{}, err
	}

	// Calculate solar declination
	jd := julianDay(date)
	n := jd - 2451545.0
	L := math.Mod(280.460+0.9856474*n, 360.0)
	g := math.Mod(357.528+0.9856003*n, 360.0) * math.Pi / 180.0

	lambda := (L + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)) * math.Pi / 180.0
	epsilon := 23.439 * math.Pi / 180.0
	declination := math.Asin(math.Sin(epsilon) * math.Sin(lambda))

	// Calculate hour angle at sunrise
	latRad := lat * math.Pi / 180.0
	cosHourAngle := -math.Tan(latRad) * math.Tan(declination)

	// Check for polar day/night
	if cosHourAngle > 1.0 || cosHourAngle < -1.0 {
		return time.Time{}, fmt.Errorf("no sunrise at this location on this date (polar day/night)")
	}

	hourAngle := math.Acos(cosHourAngle) * 180.0 / math.Pi

	// Convert to UTC time
	sunriseUTC := 12.0 - hourAngle/15.0 - lon/15.0
	hours := int(sunriseUTC)
	minutes := int((sunriseUTC - float64(hours)) * 60)

	return time.Date(date.Year(), date.Month(), date.Day(), hours, minutes, 0, 0, time.UTC), nil
}

// SunsetTime calculates the sunset time at a location on a given date.
//
// Returns the time of sunset in UTC.
//
// Example:
//
//	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
//	sunset, err := helpers.SunsetTime(45.5152, -122.6784, date)
//	fmt.Printf("Sunset: %s UTC\n", sunset.Format("15:04:05"))
func SunsetTime(lat, lon float64, date time.Time) (time.Time, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return time.Time{}, err
	}

	// Calculate solar declination
	jd := julianDay(date)
	n := jd - 2451545.0
	L := math.Mod(280.460+0.9856474*n, 360.0)
	g := math.Mod(357.528+0.9856003*n, 360.0) * math.Pi / 180.0

	lambda := (L + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)) * math.Pi / 180.0
	epsilon := 23.439 * math.Pi / 180.0
	declination := math.Asin(math.Sin(epsilon) * math.Sin(lambda))

	// Calculate hour angle at sunset
	latRad := lat * math.Pi / 180.0
	cosHourAngle := -math.Tan(latRad) * math.Tan(declination)

	// Check for polar day/night
	if cosHourAngle > 1.0 || cosHourAngle < -1.0 {
		return time.Time{}, fmt.Errorf("no sunset at this location on this date (polar day/night)")
	}

	hourAngle := math.Acos(cosHourAngle) * 180.0 / math.Pi

	// Convert to UTC time
	sunsetUTC := 12.0 + hourAngle/15.0 - lon/15.0
	hours := int(sunsetUTC)
	minutes := int((sunsetUTC - float64(hours)) * 60)

	return time.Date(date.Year(), date.Month(), date.Day(), hours, minutes, 0, 0, time.UTC), nil
}

// IsDaytime checks if it's daytime at a given location and time.
//
// Returns true if the sun is above the horizon.
//
// Example:
//
//	daytime, err := helpers.IsDaytime(45.5152, -122.6784, time.Now())
//	if daytime {
//	    fmt.Println("It's daytime!")
//	}
func IsDaytime(lat, lon float64, t time.Time) (bool, error) {
	pos, err := CalculateSunPosition(lat, lon, t)
	if err != nil {
		return false, err
	}
	return pos.Elevation > 0, nil
}

// julianDay calculates the Julian day number for a given time.
func julianDay(t time.Time) float64 {
	utc := t.UTC()
	year := utc.Year()
	month := int(utc.Month())
	day := utc.Day()

	// Adjust for January and February
	if month <= 2 {
		year--
		month += 12
	}

	a := year / 100
	b := 2 - a + a/4

	jd := float64(int(365.25*float64(year+4716))) +
		float64(int(30.6001*float64(month+1))) +
		float64(day+b) - 1524.5

	// Add time of day
	hours := float64(utc.Hour()) + float64(utc.Minute())/60.0 + float64(utc.Second())/3600.0
	jd += hours / 24.0

	return jd
}

// SolarNoon calculates the time of solar noon (when the sun is highest).
//
// Returns the time of solar noon in UTC.
//
// Example:
//
//	date := time.Date(2023, 6, 21, 0, 0, 0, 0, time.UTC)
//	noon, err := helpers.SolarNoon(-122.6784, date)
//	fmt.Printf("Solar noon: %s UTC\n", noon.Format("15:04:05"))
func SolarNoon(lon float64, date time.Time) (time.Time, error) {
	if lon < -180 || lon > 180 {
		return time.Time{}, fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", lon)
	}

	// Solar noon is at 12:00 UTC + longitude correction
	noonUTC := 12.0 - lon/15.0

	hours := int(noonUTC)
	if hours < 0 {
		hours += 24
	}
	if hours >= 24 {
		hours -= 24
	}

	minutes := int((noonUTC - float64(int(noonUTC))) * 60)
	if minutes < 0 {
		minutes += 60
		hours--
	}

	return time.Date(date.Year(), date.Month(), date.Day(), hours, minutes, 0, 0, time.UTC), nil
}
