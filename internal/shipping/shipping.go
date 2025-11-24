package shipping

import "encoding/gob"

func init() {
	gob.Register(&ShippingQuotation{})
	gob.Register(&ShippingRequest{})
	gob.Register(&Coordinates{})
	gob.Register(&Address{})
	gob.Register(ServiceType(0))
}

type Coordinates struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type Location struct {
	Coordinates     Coordinates `json:"coordinates"`
	Address         string      `json:"address"`
	OriginalAddress Address     `json:"original_address"`
	Contact         Contact     `json:"contact"`
}

type Contact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Package struct {
	Dimensions  map[string]string `json:"dimensions,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Weight      string            `json:"weight"`
	Description string            `json:"description,omitempty"`
	Value       string            `json:"value,omitempty"`
}

type ShippingRequest struct {
	Package          Package        `json:"package"`
	Options          map[string]any `json:"options,omitempty"`
	PickupLocation   Location       `json:"pickup_location"`
	DeliveryLocation Location       `json:"delivery_location"`
	ScheduledAt      string         `json:"scheduled_at,omitempty"`
	ServiceType      ServiceType    `json:"service_type,omitempty"`
}

type ShippingQuotation struct {
	Metadata     map[string]any `json:"metadata,omitempty"`
	ID           string         `json:"id,omitempty"`
	Currency     string         `json:"currency"`
	ServiceType  ServiceType    `json:"service_type"`
	ExpiresAt    string         `json:"expires_at,omitempty"`
	Fee          float64        `json:"fee"`
	DistanceKm   float64        `json:"distance_km"`
	EstimatedETA int            `json:"estimated_eta"`
}

type ShippingOrder struct {
	TrackingInfo map[string]any    `json:"tracking_info,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	Quotation    ShippingQuotation `json:"quotation"`
}

type Features struct {
	RealTimeTracking    bool `json:"real_time_tracking"`
	RouteOptimization   bool `json:"route_optimization"`
	ScheduledDelivery   bool `json:"scheduled_delivery"`
	SpecialRequests     bool `json:"special_requests"`
	MultipleStops       bool `json:"multiple_stops"`
	WeightBasedPricing  bool `json:"weight_based_pricing"`
	Insurance           bool `json:"insurance"`
	ProofOfDelivery     bool `json:"proof_of_delivery"`
	CashOnDelivery      bool `json:"cash_on_delivery"`
	ContactlessDelivery bool `json:"contactless_delivery"`
}

type ServiceCapabilities struct {
	Metadata          map[string]any `json:"metadata,omitempty"`
	Provider          string         `json:"provider"`
	APIVersion        string         `json:"api_version"`
	SupportedServices []ServiceType  `json:"supported_services"`
	Coverage          []string       `json:"coverage"`
	Features          Features       `json:"features"`
}

type IShippingService interface {
	Enum() ShippingService
	GetCapabilities() (*ServiceCapabilities, error)
	GetQuotation(ShippingRequest) (*ShippingQuotation, error)
	CreateOrder(ShippingRequest) (*ShippingOrder, error)
	GetOrderStatus(string) (*ShippingOrder, error)
	CancelOrder(string) error
	GetBusinessLocation() *Location
}
