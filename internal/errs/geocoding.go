package errs

import "errors"

var (
	ErrGeocodingNilRequest       = errors.New("[GEOCODING]: Request cannot be nil")
	ErrGeocodingNilLocation      = errors.New("[GEOCODING]: Location cannot be nil")
	ErrGeocodingEmptyAddress     = errors.New("[GEOCODING]: Address cannot be empty")
	ErrGeocodingEmptyCoordinates = errors.New("[GEOCODING]: Coordinates cannot be empty")
	ErrGeocodingAddressTooShort  = errors.New("[GEOCODING]: Address is too short, minimum 10 characters required")
	ErrGeocodingInvalidLocation  = errors.New("[GEOCODING]: Location must have either valid coordinates or an address")
)
