package geocoding

//go:generate go tool stringer -type=GeocodingService -trimprefix=GEOCODING_SERVICE_

import (
	"cchoice/internal/errs"
	"fmt"
	"strings"
)

type GeocodingService int

const (
	GEOCODING_SERVICE_UNDEFINED GeocodingService = iota
	GEOCODING_SERVICE_GOOGLEMAPS
)

func ParseGeocodingServiceToEnum(gs string) GeocodingService {
	switch strings.ToUpper(gs) {
	case GEOCODING_SERVICE_GOOGLEMAPS.String():
		return GEOCODING_SERVICE_GOOGLEMAPS
	default:
		panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUndefinedService, gs))
	}
}
