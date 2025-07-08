package paymongo

import "cchoice/internal/payments"

type CheckoutSession struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Attributes CheckoutSessionAttributes `json:"attributes"`
}

type CheckoutSessionAttributes struct {
	Billing            payments.Billing    `json:"billing"`
	CheckoutURL        string              `json:"checkout_url"`
	ClientKey          string              `json:"client_key"`
	Description        string              `json:"description"`
	LineItems          []payments.LineItem `json:"line_items"`
	Livemode           bool                `json:"livemode"`
	Merchant           string              `json:"merchant"`
	Payments           []Payments          `json:"payments"`
	PaymentIntent      PaymentIntent       `json:"payment_intent"`
	PaymentMethodTypes []string            `json:"payment_method_types"`
	ReferenceNumber    string              `json:"reference_number"`
	SendEmailReceipt   bool                `json:"send_email_receipt"`
	ShowDescription    bool                `json:"show_description"`
	ShowLineItems      bool                `json:"show_line_items"`
	Status             string              `json:"status"`
	SuccessURL         string              `json:"success_url"`
	CreatedAt          int                 `json:"created_at"`
	UpdatedAt          int                 `json:"updated_at"`
	Metadata           Metadata            `json:"metadata"`
}

type PaymentAttributes struct {
	AccessURL               string           `json:"access_url"`
	Amount                  int              `json:"amount"`
	BalanceTransactionID    string           `json:"balance_transaction_id"`
	Billing                 payments.Billing `json:"billing"`
	Currency                string           `json:"currency"`
	Description             string           `json:"description"`
	Disputed                bool             `json:"disputed"`
	ExternalReferenceNumber any              `json:"external_reference_number"`
	Fee                     int              `json:"fee"`
	ForeignFee              int              `json:"foreign_fee"`
	Livemode                bool             `json:"livemode"`
	NetAmount               int              `json:"net_amount"`
	Origin                  string           `json:"origin"`
	PaymentIntentID         string           `json:"payment_intent_id"`
	Payout                  any              `json:"payout"`
	Source                  Source           `json:"source"`
	StatementDescriptor     string           `json:"statement_descriptor"`
	Status                  string           `json:"status"`
	TaxAmount               int              `json:"tax_amount"`
	Metadata                Metadata         `json:"metadata"`
	Refunds                 []any            `json:"refunds"`
	Taxes                   []Taxes          `json:"taxes"`
	AvailableAt             int              `json:"available_at"`
	CreatedAt               int              `json:"created_at"`
	CreditedAt              int              `json:"credited_at"`
	PaidAt                  int              `json:"paid_at"`
	UpdatedAt               int              `json:"updated_at"`
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
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Inclusive bool   `json:"inclusive"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
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
	Amount               int                  `json:"amount"`
	CaptureType          string               `json:"capture_type"`
	ClientKey            string               `json:"client_key"`
	Currency             string               `json:"currency"`
	Description          string               `json:"description"`
	Livemode             bool                 `json:"livemode"`
	StatementDescriptor  string               `json:"statement_descriptor"`
	Status               string               `json:"status"`
	LastPaymentError     any                  `json:"last_payment_error"`
	PaymentMethodAllowed []string             `json:"payment_method_allowed"`
	Payments             []Payments           `json:"payments"`
	NextAction           any                  `json:"next_action"`
	PaymentMethodOptions PaymentMethodOptions `json:"payment_method_options"`
	Metadata             Metadata             `json:"metadata"`
	SetupFutureUsage     any                  `json:"setup_future_usage"`
	CreatedAt            int                  `json:"created_at"`
	UpdatedAt            int                  `json:"updated_at"`
}

type PaymentIntent struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Attributes PaymentIntentAttributes `json:"attributes"`
}
