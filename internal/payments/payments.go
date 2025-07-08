package payments

import (
	"cchoice/internal/database/queries"
)

type CreateCheckoutSessionPayload any

type CreateCheckoutSessionResponse interface {
	ToLineItems() []*queries.CreateCheckoutLineItemParams
	ToCheckout(PaymentGateway) *queries.CreateCheckoutParams
}

type GetAvailablePaymentMethodsResponse interface {
	ToPaymentMethods() []PaymentMethod
}

type PaymentGateway interface {
	GatewayName() string
	GetAuth() string
	GetAvailablePaymentMethods() (GetAvailablePaymentMethodsResponse, error)
	CreateCheckoutSession(CreateCheckoutSessionPayload) (CreateCheckoutSessionResponse, error)
}
