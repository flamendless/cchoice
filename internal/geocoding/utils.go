package geocoding

import (
	"errors"
	"strings"
)

func ValidateAddress(address string) error {
	if address == "" {
		return errors.New("address cannot be empty")
	}

	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New("address cannot be empty after trimming whitespace")
	}

	if len(address) < 10 {
		return errors.New("address is too short, minimum 10 characters required")
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
		return nil, errors.New("latitude and longitude cannot be empty")
	}

	lat = strings.TrimSpace(lat)
	lng = strings.TrimSpace(lng)

	if lat == "" || lng == "" {
		return nil, errors.New("latitude and longitude cannot be empty after trimming")
	}

	return &Coordinates{
		Lat: lat,
		Lng: lng,
	}, nil
}

func IsValidCoordinates(coords Coordinates) bool {
	return coords.Lat != "" && coords.Lng != ""
}
