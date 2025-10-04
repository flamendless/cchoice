package googlemaps

type GoogleMapsGeocodeResponse struct {
	Status       string             `json:"status"`
	ErrorMessage string             `json:"error_message,omitempty"`
	Results      []GoogleMapsResult `json:"results"`
}

type GoogleMapsResult struct {
	FormattedAddress  string                       `json:"formatted_address"`
	PlaceID           string                       `json:"place_id"`
	AddressComponents []GoogleMapsAddressComponent `json:"address_components"`
	Types             []string                     `json:"types"`
	Geometry          GoogleMapsGeometry           `json:"geometry"`
}

type GoogleMapsAddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type GoogleMapsGeometry struct {
	Viewport     *GoogleMapsBounds `json:"viewport,omitempty"`
	Bounds       *GoogleMapsBounds `json:"bounds,omitempty"`
	LocationType string            `json:"location_type"`
	Location     GoogleMapsLatLng  `json:"location"`
}

type GoogleMapsLatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GoogleMapsBounds struct {
	Northeast GoogleMapsLatLng `json:"northeast"`
	Southwest GoogleMapsLatLng `json:"southwest"`
}
