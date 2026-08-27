// Package geo provides the haversine distance calculation used for
// geofencing. Legacy code duplicated this formula verbatim in both
// PresensiController and AbsensiController (LOGIC_SPEC.md §4) -- kept here
// as the single source of truth instead.
package geo

import "math"

const earthRadiusMeters = 6371000.0

// DistanceMeters returns the great-circle distance between two lat/lng
// points, in meters.
func DistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
