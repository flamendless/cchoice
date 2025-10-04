package shipping

import (
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"fmt"
)

type GeocodingHelper struct {
	geocoder geocoding.IGeocoder
}

func NewGeocodingHelper(geocoder geocoding.IGeocoder) *GeocodingHelper {
	return &GeocodingHelper{
		geocoder: geocoder,
	}
}

func (gh *GeocodingHelper) EnsureCoordinates(request *ShippingRequest) error {
	if request == nil {
		return errs.ErrGeocodingNilRequest
	}

	if err := gh.ensureLocationCoordinates(&request.PickupLocation, "pickup"); err != nil {
		return err
	}

	if err := gh.ensureLocationCoordinates(&request.DeliveryLocation, "delivery"); err != nil {
		return err
	}

	return nil
}

func (gh *GeocodingHelper) ensureLocationCoordinates(location *Location, locationType string) error {
	if location == nil {
		return fmt.Errorf("%s location cannot be nil", locationType)
	}

	if location.Coordinates.Lat != "" && location.Coordinates.Lng != "" {
		return nil
	}

	if location.Address == "" {
		return fmt.Errorf("%s location address cannot be empty for geocoding", locationType)
	}

	coordinates, err := gh.geocoder.GeocodeShippingAddress(location.Address)
	if err != nil {
		return fmt.Errorf("failed to geocode %s address '%s': %w", locationType, location.Address, err)
	}

	location.Coordinates = Coordinates{
		Lat: coordinates.Lat,
		Lng: coordinates.Lng,
	}
	return nil
}

func (gh *GeocodingHelper) GeocodeAddress(address string) (*Coordinates, error) {
	if address == "" {
		return nil, errs.ErrGeocodingEmptyAddress
	}

	coordinates, err := gh.geocoder.GeocodeShippingAddress(address)
	if err != nil {
		return nil, err
	}

	return &Coordinates{
		Lat: coordinates.Lat,
		Lng: coordinates.Lng,
	}, nil
}

func (gh *GeocodingHelper) ReverseGeocode(coordinates Coordinates) (string, error) {
	if coordinates.Lat == "" || coordinates.Lng == "" {
		return "", errs.ErrGeocodingEmptyCoordinates
	}

	req := geocoding.ReverseGeocodeRequest{
		Coordinates: geocoding.Coordinates{
			Lat: coordinates.Lat,
			Lng: coordinates.Lng,
		},
		Language: "en",
	}

	result, err := gh.geocoder.ReverseGeocode(req)
	if err != nil {
		return "", fmt.Errorf("failed to reverse geocode coordinates: %w", err)
	}

	return result.FormattedAddress, nil
}

func (gh *GeocodingHelper) ValidateShippingRequest(request *ShippingRequest) error {
	if request == nil {
		return errs.ErrGeocodingNilRequest
	}

	if err := gh.validateLocation(&request.PickupLocation, "pickup"); err != nil {
		return err
	}

	if err := gh.validateLocation(&request.DeliveryLocation, "delivery"); err != nil {
		return err
	}

	return nil
}

func (gh *GeocodingHelper) validateLocation(location *Location, locationType string) error {
	if location == nil {
		return fmt.Errorf("%s location cannot be nil", locationType)
	}

	hasCoordinates := location.Coordinates.Lat != "" && location.Coordinates.Lng != ""
	hasAddress := location.Address != ""

	if !hasCoordinates && !hasAddress {
		return fmt.Errorf("%w: %s location", errs.ErrGeocodingInvalidLocation, locationType)
	}

	return nil
}

func GeocodeShippingLocation(geocoder geocoding.IGeocoder, location *Location) error {
	if location == nil {
		return errs.ErrGeocodingNilLocation
	}

	if location.Address == "" {
		return errs.ErrGeocodingEmptyAddress
	}

	if location.Coordinates.Lat != "" && location.Coordinates.Lng != "" {
		return nil
	}

	coordinates, err := geocoder.GeocodeShippingAddress(location.Address)
	if err != nil {
		return fmt.Errorf("failed to geocode address '%s': %w", location.Address, err)
	}

	location.Coordinates = Coordinates{
		Lat: coordinates.Lat,
		Lng: coordinates.Lng,
	}
	return nil
}

func GeocodeShippingRequest(geocoder geocoding.IGeocoder, request *ShippingRequest) error {
	if request == nil {
		return errs.ErrGeocodingNilRequest
	}

	if err := GeocodeShippingLocation(geocoder, &request.PickupLocation); err != nil {
		return fmt.Errorf("failed to geocode pickup location: %w", err)
	}

	if err := GeocodeShippingLocation(geocoder, &request.DeliveryLocation); err != nil {
		return fmt.Errorf("failed to geocode delivery location: %w", err)
	}

	return nil
}
