package googlemaps

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GoogleMapsGeocoder struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	region     string
}

func MustInit() *GoogleMapsGeocoder {
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
		region: "PH",
	}
}

func (g *GoogleMapsGeocoder) SetRegion(region string) {
	g.region = region
}

func (g *GoogleMapsGeocoder) Geocode(req geocoding.GeocodeRequest) (*geocoding.GeocodeResponse, error) {
	if req.Address == "" {
		return nil, errs.ErrInvalidRequest
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

	resp, err := g.httpClient.Get(requestURL)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp GoogleMapsGeocodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if err := g.checkStatus(apiResp.Status); err != nil {
		return nil, err
	}

	if len(apiResp.Results) == 0 {
		return nil, errs.ErrNoResults
	}

	result := apiResp.Results[0]
	return g.convertToGeocodeResponse(result), nil
}

func (g *GoogleMapsGeocoder) ReverseGeocode(req geocoding.ReverseGeocodeRequest) (*geocoding.GeocodeResponse, error) {
	if req.Coordinates.Lat == "" || req.Coordinates.Lng == "" {
		return nil, errs.ErrInvalidRequest
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
		return nil, err
	}

	var apiResp GoogleMapsGeocodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if err := g.checkStatus(apiResp.Status); err != nil {
		return nil, err
	}

	if len(apiResp.Results) == 0 {
		return nil, errs.ErrNoResults
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
		return errs.ErrNoResults
	case "OVER_QUERY_LIMIT":
		return errs.ErrQuotaExceeded
	case "REQUEST_DENIED":
		return errs.ErrRequestDenied
	case "INVALID_REQUEST":
		return errs.ErrInvalidRequest
	default:
		return fmt.Errorf("%w: %s", errs.ErrUnknownError, status)
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
