package geocoding

type GeocodeRequest struct {
	ComponentFilter map[string]string `json:"components,omitempty"`
	Address         string            `json:"address"`
	Region          string            `json:"region,omitempty"`
	Language        string            `json:"language,omitempty"`
}

type GeocodeResponse struct {
	ViewportBounds    *ViewportBounds    `json:"viewport,omitempty"`
	Coordinates       Coordinates        `json:"coordinates"`
	FormattedAddress  string             `json:"formatted_address"`
	PlaceID           string             `json:"place_id,omitempty"`
	Types             []string           `json:"types,omitempty"`
	AddressComponents []AddressComponent `json:"address_components,omitempty"`
}

type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type ViewportBounds struct {
	Northeast Coordinates `json:"northeast"`
	Southwest Coordinates `json:"southwest"`
}

type ReverseGeocodeRequest struct {
	Coordinates Coordinates `json:"coordinates"`
	Language    string      `json:"language,omitempty"`
	ResultTypes []string    `json:"result_types,omitempty"`
}

type IGeocoder interface {
	Geocode(req GeocodeRequest) (*GeocodeResponse, error)
	ReverseGeocode(req ReverseGeocodeRequest) (*GeocodeResponse, error)
	GeocodeShippingAddress(address string) (*Coordinates, error)
	SetRegion(region string)
}

type Coordinates struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}
