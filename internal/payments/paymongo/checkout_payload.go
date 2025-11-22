package paymongo

import (
	"cchoice/internal/payments"
)

type CreateCheckoutSessionAttr struct {
	Billing             payments.Billing    `json:"billing"`
	CancelURL           string              `json:"cancel_url"`
	Description         string              `json:"description"`
	ReferenceNumber     string              `json:"reference_number"`
	StatementDescriptor string              `json:"statement_descriptor"`
	SuccessURL          string              `json:"success_url"`
	LineItems           []payments.LineItem `json:"line_items"`
	PaymentMethodTypes  []string            `json:"payment_method_types"`
	SendEmailReceipt    bool                `json:"send_email_receipt"`
	ShowDescription     bool                `json:"show_description"`
	ShowLineItems       bool                `json:"show_line_items"`
}

type CreateCheckoutSessionData struct {
	Attributes CreateCheckoutSessionAttr `json:"attributes"`
}

type CreateCheckoutSessionPayload struct {
	Data CreateCheckoutSessionData `json:"data"`
}

var _ payments.CreateCheckoutSessionPayload = (*CreateCheckoutSessionPayload)(nil)
