package utils

import (
	"cchoice/internal/constants"
	"cchoice/internal/types"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
)

func HaversineDistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return constants.EarthRadiusMeters * c
}

func IsWithinRadius(lat1, lng1, lat2, lng2 float64, radiusMeters int) bool {
	if radiusMeters <= 0 {
		return false
	}
	return HaversineDistanceMeters(lat1, lng1, lat2, lng2) <= float64(radiusMeters)
}

func FormatLocation(location string) string {
	if location == "" {
		return ""
	}
	var sl types.Location
	if err := json.Unmarshal([]byte(location), &sl); err != nil {
		return location
	}
	return fmt.Sprintf("(%f,%f)", sl.Lat, sl.Lng)
}

func ParseLocation(locationJSON sql.NullString) (lat, lng float64, ok bool) {
	if !locationJSON.Valid || locationJSON.String == "" {
		return 0, 0, false
	}
	var loc types.Location
	if err := json.Unmarshal([]byte(locationJSON.String), &loc); err != nil {
		return 0, 0, false
	}
	return loc.Lat, loc.Lng, true
}
