package payments

import (
	"cchoice/internal/enums"
)

type PayMongoCreateCheckoutSessionAttr struct {
	Billing             Billing               `json:"billing"`
	CancelURL           string                `json:"cancel_url"`
	Description         string                `json:"description"`
	ReferenceNumber     string                `json:"reference_number"`
	StatementDescriptor string                `json:"statement_descriptor"`
	SuccessURL          string                `json:"success_url"`
	LineItems           []LineItem            `json:"line_items"`
	PaymentMethodTypes  []enums.PaymentMethod `json:"payment_method_types"`
	SendEmailReceipt    bool                  `json:"send_email_receipt"`
	ShowDescription     bool                  `json:"show_description"`
	ShowLineItems       bool                  `json:"show_line_items"`
}

type PayMongoCreateCheckoutSessionData struct {
	Attributes PayMongoCreateCheckoutSessionAttr `json:"attributes"`
}

type PayMongoCreateCheckoutSessionPayload struct {
	Data PayMongoCreateCheckoutSessionData `json:"data"`
}

type PayMongoCreateCheckoutSessionResponse struct {
}

var _ createCheckoutSessionPayload = (*PayMongoCreateCheckoutSessionPayload)(nil)
var _ createCheckoutSessionResponse = (*PayMongoCreateCheckoutSessionResponse)(nil)
