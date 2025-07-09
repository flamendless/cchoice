package payments

import (
	"cchoice/internal/database/queries"
	"net/http"
)

type CreateCheckoutSessionPayload any

type CreateCheckoutSessionResponse interface {
	ToLineItems() []*queries.CreateCheckoutLineItemParams
	ToCheckout(IPaymentGateway) *queries.CreateCheckoutParams
}

type GetAvailablePaymentMethodsResponse interface {
	ToPaymentMethods() []PaymentMethod
}

type IPaymentGateway interface {
	GatewayEnum() PaymentGateway
	GetAuth() string
	GetAvailablePaymentMethods() (GetAvailablePaymentMethodsResponse, error)
	CreateCheckoutSession(CreateCheckoutSessionPayload) (CreateCheckoutSessionResponse, error)
	CheckoutHanlder(http.ResponseWriter, *http.Request) error
}
