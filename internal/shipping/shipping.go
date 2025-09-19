package shipping

type Coordinates struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type Location struct {
	Coordinates Coordinates `json:"coordinates"`
	Address     string      `json:"address"`
	Contact     Contact     `json:"contact,omitempty"`
}

type Contact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Package struct {
	Weight      string            `json:"weight"`
	Dimensions  map[string]string `json:"dimensions,omitempty"`
	Description string            `json:"description,omitempty"`
	Value       string            `json:"value,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

type ShippingRequest struct {
	PickupLocation   Location       `json:"pickup_location"`
	DeliveryLocation Location       `json:"delivery_location"`
	Package          Package        `json:"package"`
	ScheduledAt      string         `json:"scheduled_at,omitempty"`
	ServiceType      string         `json:"service_type,omitempty"`
	Options          map[string]any `json:"options,omitempty"`
}

type ShippingQuotation struct {
	ID           string         `json:"id,omitempty"`
	Currency     string         `json:"currency"`
	Fee          float64        `json:"fee"`
	DistanceKm   float64        `json:"distance_km"`
	EstimatedETA int            `json:"estimated_eta"`
	ServiceType  string         `json:"service_type"`
	ExpiresAt    string         `json:"expires_at,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type ShippingOrder struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	Quotation    ShippingQuotation `json:"quotation"`
	TrackingInfo map[string]any    `json:"tracking_info,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
}

// Features represents the capabilities supported by a shipping service
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
	SupportedServices []string       `json:"supported_services"`
	Coverage          []string       `json:"coverage"`
	Features          Features       `json:"features"`
	Provider          string         `json:"provider"`
	APIVersion        string         `json:"api_version"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

type IShippingService interface {
	Enum() ShippingService
	GetCapabilities() (*ServiceCapabilities, error)
	GetQuotation(ShippingRequest) (*ShippingQuotation, error)
	CreateOrder(ShippingRequest) (*ShippingOrder, error)
	GetOrderStatus(string) (*ShippingOrder, error)
	CancelOrder(string) error
}
