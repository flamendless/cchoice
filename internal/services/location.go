package services

import (
	"database/sql"
	"fmt"

	"cchoice/internal/types"
	"cchoice/internal/utils"
)

type LocationResult struct {
	InShop          *bool
	DistanceMeters  float64
	LocationDisplay string
}

func ComputeLocation(
	attendanceLocation sql.NullString,
	sessionLocation sql.NullString,
	shop types.Location,
) LocationResult {
	var lat, lng float64
	var ok bool

	if attendanceLocation.Valid {
		lat, lng, ok = utils.ParseLocation(attendanceLocation)
	}

	if !ok && sessionLocation.Valid {
		lat, lng, ok = utils.ParseLocation(sessionLocation)
	}

	result := LocationResult{
		LocationDisplay: "unable to get location",
	}

	if !ok {
		return result
	}

	result.LocationDisplay = fmt.Sprintf("%.4f, %.4f", lat, lng)

	if shop.RadiusMeters > 0 {
		in := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
		result.InShop = &in
		result.DistanceMeters = utils.HaversineDistanceMeters(lat, lng, shop.Lat, shop.Lng)
	}

	return result
}
