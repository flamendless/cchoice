package paymongo

import "cchoice/internal/payments"

type CheckoutSession struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Attributes CheckoutSessionAttributes `json:"attributes"`
}

type CheckoutSessionAttributes struct {
	Billing            payments.Billing    `json:"billing"`
	Metadata           Metadata            `json:"metadata"`
	CheckoutURL        string              `json:"checkout_url"`
	ClientKey          string              `json:"client_key"`
	Description        string              `json:"description"`
	Merchant           string              `json:"merchant"`
	ReferenceNumber    string              `json:"reference_number"`
	Status             string              `json:"status"`
	SuccessURL         string              `json:"success_url"`
	PaymentIntent      PaymentIntent       `json:"payment_intent"`
	LineItems          []payments.LineItem `json:"line_items"`
	Payments           []Payments          `json:"payments"`
	PaymentMethodTypes []string            `json:"payment_method_types"`
	CreatedAt          int                 `json:"created_at"`
	UpdatedAt          int                 `json:"updated_at"`
	Livemode           bool                `json:"livemode"`
	SendEmailReceipt   bool                `json:"send_email_receipt"`
	ShowDescription    bool                `json:"show_description"`
	ShowLineItems      bool                `json:"show_line_items"`
}

type PaymentAttributes struct {
	ExternalReferenceNumber any              `json:"external_reference_number"`
	Payout                  any              `json:"payout"`
	Billing                 payments.Billing `json:"billing"`
	Source                  Source           `json:"source"`
	Metadata                Metadata         `json:"metadata"`
	AccessURL               string           `json:"access_url"`
	BalanceTransactionID    string           `json:"balance_transaction_id"`
	Currency                string           `json:"currency"`
	Description             string           `json:"description"`
	Origin                  string           `json:"origin"`
	PaymentIntentID         string           `json:"payment_intent_id"`
	StatementDescriptor     string           `json:"statement_descriptor"`
	Status                  string           `json:"status"`
	Refunds                 []any            `json:"refunds"`
	Taxes                   []Taxes          `json:"taxes"`
	Amount                  int              `json:"amount"`
	Fee                     int              `json:"fee"`
	ForeignFee              int              `json:"foreign_fee"`
	NetAmount               int              `json:"net_amount"`
	TaxAmount               int              `json:"tax_amount"`
	AvailableAt             int              `json:"available_at"`
	CreatedAt               int              `json:"created_at"`
	CreditedAt              int              `json:"credited_at"`
	PaidAt                  int              `json:"paid_at"`
	UpdatedAt               int              `json:"updated_at"`
	Disputed                bool             `json:"disputed"`
	Livemode                bool             `json:"livemode"`
}

type Source struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Brand   string `json:"brand"`
	Country string `json:"country"`
	Last4   string `json:"last4"`
}

type Metadata struct {
	CustomerNumber string `json:"customer_number"`
	Remarks        string `json:"remarks"`
	Notes          string `json:"notes"`
}

type Taxes struct {
	Currency  string `json:"currency"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Amount    int    `json:"amount"`
	Inclusive bool   `json:"inclusive"`
}

type Payments struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes PaymentAttributes `json:"attributes"`
}

type Installments struct {
	Enabled bool `json:"enabled"`
}

type Card struct {
	RequestThreeDSecure string       `json:"request_three_d_secure"`
	Installments        Installments `json:"installments"`
}

type PaymentMethodOptions struct {
	Card Card `json:"card"`
}

type PaymentIntentAttributes struct {
	LastPaymentError     any                  `json:"last_payment_error"`
	NextAction           any                  `json:"next_action"`
	SetupFutureUsage     any                  `json:"setup_future_usage"`
	Metadata             Metadata             `json:"metadata"`
	CaptureType          string               `json:"capture_type"`
	ClientKey            string               `json:"client_key"`
	Currency             string               `json:"currency"`
	Description          string               `json:"description"`
	StatementDescriptor  string               `json:"statement_descriptor"`
	Status               string               `json:"status"`
	PaymentMethodAllowed []string             `json:"payment_method_allowed"`
	Payments             []Payments           `json:"payments"`
	PaymentMethodOptions PaymentMethodOptions `json:"payment_method_options"`
	Amount               int                  `json:"amount"`
	CreatedAt            int                  `json:"created_at"`
	UpdatedAt            int                  `json:"updated_at"`
	Livemode             bool                 `json:"livemode"`
}

type PaymentIntent struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Attributes PaymentIntentAttributes `json:"attributes"`
}
