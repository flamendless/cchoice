package payments

type CreateCheckoutSessionPayload any

type CreateCheckoutSessionResponse any

type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type Billing struct {
	Address Address `json:"address"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
}

type LineItem struct {
	Currency    string   `json:"currency"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	Images      []string `json:"images"`
	Amount      int32    `json:"amount"`
	Quantity    int32    `json:"quantity"`
}

type PaymentGateway interface {
	GatewayName() string
	GetAuth() string
	CreateCheckoutSession(payload CreateCheckoutSessionPayload) (CreateCheckoutSessionResponse, error)
}
