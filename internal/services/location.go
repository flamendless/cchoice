package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"cchoice/internal/logs"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"github.com/alexedwards/scs/v2"
)

type LocationResult struct {
	InShop          *bool
	DistanceMeters  float64
	LocationDisplay string
}

type LocationService struct {
	shopLocation types.Location
}

func NewLocationService(shopLocation types.Location) *LocationService {
	return &LocationService{
		shopLocation: shopLocation,
	}
}

func (s *LocationService) CheckShopRadius(
	ctx context.Context,
	sessionManager *scs.SessionManager,
	attendanceInLocation, attendanceOutLocation sql.NullString,
) (inShop, outShop *bool) {
	if s.shopLocation.RadiusMeters <= 0 {
		return nil, nil
	}

	if attendanceInLocation.Valid {
		if lat, lng, ok := utils.ParseLocation(attendanceInLocation); ok {
			b := utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
			inShop = &b
		}
	}
	if attendanceOutLocation.Valid {
		if lat, lng, ok := utils.ParseLocation(attendanceOutLocation); ok {
			b := utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
			outShop = &b
		}
	}

	sessionLocation := s.GetSessionLocation(ctx, sessionManager)
	if inShop == nil && sessionLocation.Valid {
		if lat, lng, ok := utils.ParseLocation(sessionLocation); ok {
			b := utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
			inShop = &b
		}
	}
	if outShop == nil && sessionLocation.Valid {
		if lat, lng, ok := utils.ParseLocation(sessionLocation); ok {
			b := utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
			outShop = &b
		}
	}

	return inShop, outShop
}

func (s *LocationService) ComputeLocationDisplay(
	ctx context.Context,
	sessionManager *scs.SessionManager,
) (locationDisplay string, distanceMeters float64) {
	locationDisplay = "unable to get location"
	distanceMeters = 0.0

	sessionLocation := s.GetSessionLocation(ctx, sessionManager)
	if !sessionLocation.Valid {
		return locationDisplay, distanceMeters
	}

	if lat, lng, ok := utils.ParseLocation(sessionLocation); ok {
		locationDisplay = fmt.Sprintf("%.4f, %.4f", lat, lng)
		if s.shopLocation.Lat != 0 && s.shopLocation.Lng != 0 {
			distanceMeters = utils.HaversineDistanceMeters(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng)
		}
	} else {
		locationDisplay = sessionLocation.String
	}

	return locationDisplay, distanceMeters
}

func (s *LocationService) ComputeLocationFromRequest(r *http.Request) (lat, lng float64, err error) {
	var latStr, lngStr string

	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return 0, 0, fmt.Errorf("invalid body: %w", err)
		}
		latStr = strconv.FormatFloat(body.Lat, 'f', -1, 64)
		lngStr = strconv.FormatFloat(body.Lng, 'f', -1, 64)
	} else {
		_ = r.ParseForm()
		latStr = r.PostFormValue("lat")
		lngStr = r.PostFormValue("lng")
	}

	if latStr == "" || lngStr == "" {
		return 0, 0, errors.New("lat and lng required")
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("invalid lat/lng")
	}

	return lat, lng, nil
}

func (s *LocationService) GetSessionLocation(ctx context.Context, sm *scs.SessionManager) sql.NullString {
	latVal := sm.Get(ctx, "location_lat")
	lngVal := sm.Get(ctx, "location_lng")
	if latVal == nil || lngVal == nil {
		return sql.NullString{}
	}
	lat, ok1 := latVal.(float64)
	lng, ok2 := lngVal.(float64)
	if !ok1 || !ok2 {
		return sql.NullString{}
	}
	b, _ := json.Marshal(types.Location{Lat: lat, Lng: lng})
	return sql.NullString{String: string(b), Valid: true}
}

func (s *LocationService) ComputeLocation(
	attendanceLocation sql.NullString,
	sessionLocation sql.NullString,
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

	if s.shopLocation.RadiusMeters > 0 {
		in := utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
		result.InShop = &in
		result.DistanceMeters = utils.HaversineDistanceMeters(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng)
	}

	return result
}

func (s *LocationService) Log() {
	logs.Log().Info("[LocationService] Loaded")
}

var _ IService = (*LocationService)(nil)
