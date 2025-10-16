package googlemaps

import (
	"cchoice/internal/conf"
	"cchoice/internal/database"
	dbqueries "cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"cchoice/internal/logs"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const DB_NIL = "DB is not initialized, skipping cache store"

type GoogleMapsGeocoder struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	region      string
	db          database.Service
	cacheExpiry time.Duration
}

func MustInit(db database.Service) *GoogleMapsGeocoder {
	cfg := conf.Conf()
	if cfg.GeocodingService != "googlemaps" {
		panic("'GEOCODING_SERVICE' must be 'googlemaps' to use this")
	}

	return &GoogleMapsGeocoder{
		apiKey:  cfg.GoogleMapsAPIKey,
		baseURL: "https://maps.googleapis.com/maps/api/geocode/json",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		region:      "PH",
		db:          db,
		cacheExpiry: 0,
	}
}

func MustInitWithCache(db database.Service, cacheExpiry time.Duration) *GoogleMapsGeocoder {
	cfg := conf.Conf()
	if cfg.GeocodingService != "googlemaps" {
		panic("'GEOCODING_SERVICE' must be 'googlemaps' to use this")
	}

	if cacheExpiry == 0 {
		cacheExpiry = 30 * 24 * time.Hour
	}

	return &GoogleMapsGeocoder{
		apiKey:  cfg.GoogleMapsAPIKey,
		baseURL: "https://maps.googleapis.com/maps/api/geocode/json",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		region:      "PH",
		db:          db,
		cacheExpiry: cacheExpiry,
	}
}

func (g *GoogleMapsGeocoder) SetRegion(region string) {
	g.region = region
}

func (g *GoogleMapsGeocoder) normalizeAddress(address string) string {
	normalized := strings.ToLower(strings.TrimSpace(address))
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func (g *GoogleMapsGeocoder) checkDBCache(address string) (*geocoding.GeocodeResponse, bool) {
	if g.db == nil {
		logs.Log().Warn(DB_NIL)
		return nil, false
	}

	ctx := context.Background()
	queries := g.db.GetQueries()
	normalizedAddr := g.normalizeAddress(address)

	cached, err := queries.GetGeocodingCacheByAddress(ctx, normalizedAddr)
	if err == nil {
		logs.Log().Debug("Geocoding DB cache hit",
			zap.String("address", address),
			zap.String("normalized", normalizedAddr))

		response := &geocoding.GeocodeResponse{
			Coordinates: geocoding.Coordinates{
				Lat: cached.Latitude,
				Lng: cached.Longitude,
			},
			FormattedAddress: cached.FormattedAddress,
		}

		if cached.PlaceID.Valid {
			response.PlaceID = cached.PlaceID.String
		}

		if cached.ResponseData.Valid {
			var fullResponse geocoding.GeocodeResponse
			if err := json.Unmarshal([]byte(cached.ResponseData.String), &fullResponse); err == nil {
				response = &fullResponse
			}
		}

		return response, true
	}

	if err != sql.ErrNoRows {
		logs.Log().Warn("Error checking geocoding DB cache",
			zap.Error(err),
			zap.String("address", address),
		)
	} else {
		logs.Log().Debug("Geocoding DB cache miss",
			zap.String("address", address),
			zap.String("normalized", normalizedAddr),
		)
	}

	return nil, false
}

func (g *GoogleMapsGeocoder) storeDBCache(address string, response *geocoding.GeocodeResponse) {
	if g.db == nil {
		logs.Log().Warn(DB_NIL)
		return
	}

	ctx := context.Background()
	queries := g.db.GetQueries()
	normalizedAddr := g.normalizeAddress(address)

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logs.Log().Warn("Failed to marshal geocoding response", zap.Error(err))
		return
	}

	var expiresAt sql.NullTime
	if g.cacheExpiry > 0 {
		expiresAt = sql.NullTime{
			Time:  time.Now().Add(g.cacheExpiry),
			Valid: true,
		}
	}

	placeID := sql.NullString{
		String: response.PlaceID,
		Valid:  response.PlaceID != "",
	}

	params := dbqueries.UpsertGeocodingCacheParams{
		Address:           address,
		NormalizedAddress: normalizedAddr,
		Latitude:          response.Coordinates.Lat,
		Longitude:         response.Coordinates.Lng,
		FormattedAddress:  response.FormattedAddress,
		PlaceID:           placeID,
		ResponseData: sql.NullString{
			String: string(responseJSON),
			Valid:  true,
		},
		ExpiresAt: expiresAt,
	}

	_, err = queries.UpsertGeocodingCache(ctx, params)
	if err != nil {
		logs.Log().Warn("Failed to cache geocoding result in DB",
			zap.Error(err),
			zap.String("address", address),
		)
	} else {
		logs.Log().Debug("Cached geocoding result in DB",
			zap.String("address", address),
			zap.String("normalized", normalizedAddr),
		)
	}
}

func (g *GoogleMapsGeocoder) geocodeAPI(req geocoding.GeocodeRequest) (*geocoding.GeocodeResponse, error) {
	if req.Address == "" {
		return nil, errs.ErrGMapsInvalidRequest
	}

	params := url.Values{}
	params.Set("address", req.Address)
	params.Set("key", g.apiKey)

	region := req.Region
	if region == "" {
		region = g.region
	}
	if region != "" {
		params.Set("region", region)
	}

	if req.Language != "" {
		params.Set("language", req.Language)
	}

	if len(req.ComponentFilter) > 0 {
		var components []string
		for key, value := range req.ComponentFilter {
			components = append(components, fmt.Sprintf("%s:%s", key, value))
		}
		params.Set("components", strings.Join(components, "|"))
	}

	requestURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())

	logs.Log().Debug("Making Google Maps API request", zap.String("address", req.Address))

	resp, err := g.httpClient.Get(requestURL)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(errs.ErrGMapsInvalidResponse, errs.ErrIORead, err)
	}

	var apiResp GoogleMapsGeocodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, errors.Join(errs.ErrGMapsInvalidResponse, errs.ErrJSONUnmarshal, err)
	}

	if err := g.checkStatus(apiResp.Status); err != nil {
		return nil, err
	}

	if len(apiResp.Results) == 0 {
		return nil, errs.ErrGMapsNoResults
	}

	result := apiResp.Results[0]
	return g.convertToGeocodeResponse(result), nil
}

func (g *GoogleMapsGeocoder) Geocode(req geocoding.GeocodeRequest) (*geocoding.GeocodeResponse, error) {
	if req.Address == "" {
		return nil, errs.ErrGMapsInvalidRequest
	}

	logs.Log().Info("Checking DB cache for geocoding", zap.String("address", req.Address))
	if response, found := g.checkDBCache(req.Address); found {
		return response, nil
	}

	logs.Log().Info("Making API call for geocoding", zap.String("address", req.Address))
	response, err := g.geocodeAPI(req)
	if err != nil {
		return nil, err
	}

	logs.Log().Info("Storing geocoding result in DB", zap.String("address", req.Address))
	g.storeDBCache(req.Address, response)

	return response, nil
}

func (g *GoogleMapsGeocoder) ReverseGeocode(req geocoding.ReverseGeocodeRequest) (*geocoding.GeocodeResponse, error) {
	if req.Coordinates.Lat == "" || req.Coordinates.Lng == "" {
		return nil, errs.ErrGMapsInvalidRequest
	}

	params := url.Values{}
	params.Set("latlng", fmt.Sprintf("%s,%s", req.Coordinates.Lat, req.Coordinates.Lng))
	params.Set("key", g.apiKey)

	if req.Language != "" {
		params.Set("language", req.Language)
	}

	if len(req.ResultTypes) > 0 {
		params.Set("result_type", strings.Join(req.ResultTypes, "|"))
	}

	requestURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())

	resp, err := g.httpClient.Get(requestURL)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(errs.ErrGMapsInvalidResponse, errs.ErrIORead, err)
	}

	var apiResp GoogleMapsGeocodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, errors.Join(errs.ErrGMapsInvalidResponse, errs.ErrJSONUnmarshal, err)
	}

	if err := g.checkStatus(apiResp.Status); err != nil {
		return nil, err
	}

	if len(apiResp.Results) == 0 {
		return nil, errs.ErrGMapsNoResults
	}

	result := apiResp.Results[0]
	return g.convertToGeocodeResponse(result), nil
}

func (g *GoogleMapsGeocoder) GeocodeShippingAddress(address string) (*geocoding.Coordinates, error) {
	req := geocoding.GeocodeRequest{
		Address: address,
		ComponentFilter: map[string]string{
			"country": "PH",
		},
	}

	result, err := g.Geocode(req)
	if err != nil {
		return nil, err
	}

	return &result.Coordinates, nil
}

func (g *GoogleMapsGeocoder) checkStatus(status string) error {
	switch status {
	case "OK":
		return nil
	case "ZERO_RESULTS":
		return errs.ErrGMapsNoResults
	case "OVER_QUERY_LIMIT":
		return errs.ErrGMapsQuotaExceeded
	case "REQUEST_DENIED":
		return errs.ErrGMapsRequestDenied
	case "INVALID_REQUEST":
		return errs.ErrGMapsInvalidRequest
	default:
		return fmt.Errorf("%w: %s", errs.ErrGMapsUnknownError, status)
	}
}

func (g *GoogleMapsGeocoder) convertToGeocodeResponse(result GoogleMapsResult) *geocoding.GeocodeResponse {
	response := &geocoding.GeocodeResponse{
		Coordinates: geocoding.Coordinates{
			Lat: strconv.FormatFloat(result.Geometry.Location.Lat, 'f', -1, 64),
			Lng: strconv.FormatFloat(result.Geometry.Location.Lng, 'f', -1, 64),
		},
		FormattedAddress: result.FormattedAddress,
		PlaceID:          result.PlaceID,
		Types:            result.Types,
	}

	if len(result.AddressComponents) > 0 {
		response.AddressComponents = make([]geocoding.AddressComponent, len(result.AddressComponents))
		for i, comp := range result.AddressComponents {
			response.AddressComponents[i] = geocoding.AddressComponent{
				LongName:  comp.LongName,
				ShortName: comp.ShortName,
				Types:     comp.Types,
			}
		}
	}

	if result.Geometry.Viewport != nil {
		response.ViewportBounds = &geocoding.ViewportBounds{
			Northeast: geocoding.Coordinates{
				Lat: strconv.FormatFloat(result.Geometry.Viewport.Northeast.Lat, 'f', -1, 64),
				Lng: strconv.FormatFloat(result.Geometry.Viewport.Northeast.Lng, 'f', -1, 64),
			},
			Southwest: geocoding.Coordinates{
				Lat: strconv.FormatFloat(result.Geometry.Viewport.Southwest.Lat, 'f', -1, 64),
				Lng: strconv.FormatFloat(result.Geometry.Viewport.Southwest.Lng, 'f', -1, 64),
			},
		}
	}

	return response
}

var _ geocoding.IGeocoder = (*GoogleMapsGeocoder)(nil)
