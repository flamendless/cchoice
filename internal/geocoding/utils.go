package geocoding

import (
	"cchoice/internal/errs"
	"strings"
)

func ValidateAddress(address string) error {
	if address == "" {
		return errs.ErrGeocodingEmptyAddress
	}

	address = strings.TrimSpace(address)
	if address == "" {
		return errs.ErrGeocodingEmptyAddress
	}

	if len(address) < 10 {
		return errs.ErrGeocodingAddressTooShort
	}

	return nil
}

func FormatAddressForGeocoding(address, city, province, country string) string {
	var parts []string

	if address != "" {
		parts = append(parts, strings.TrimSpace(address))
	}

	if city != "" {
		parts = append(parts, strings.TrimSpace(city))
	}

	if province != "" {
		parts = append(parts, strings.TrimSpace(province))
	}

	if country != "" {
		parts = append(parts, strings.TrimSpace(country))
	} else {
		parts = append(parts, "Philippines")
	}

	return strings.Join(parts, ", ")
}

func ParseCoordinates(lat, lng string) (*Coordinates, error) {
	if lat == "" || lng == "" {
		return nil, errs.ErrGeocodingEmptyCoordinates
	}

	lat = strings.TrimSpace(lat)
	lng = strings.TrimSpace(lng)

	if lat == "" || lng == "" {
		return nil, errs.ErrGeocodingEmptyCoordinates
	}

	return &Coordinates{
		Lat: lat,
		Lng: lng,
	}, nil
}

func IsValidCoordinates(coords Coordinates) bool {
	return coords.Lat != "" && coords.Lng != ""
}
