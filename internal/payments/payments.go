package payments

import (
	"cchoice/internal/database/queries"
	"net/http"
)

type CreateCheckoutSessionPayload any

type CreateCheckoutSessionResponse interface {
	ToLineItems(int64) []*queries.CreateCheckoutLineParams
	ToCheckoutPayment(IPaymentGateway) *queries.CreateCheckoutPaymentParams
}

type GetAvailablePaymentMethodsResponse interface {
	ToPaymentMethods() []PaymentMethod
}

type IPaymentGateway interface {
	GatewayEnum() PaymentGateway
	GetAuth() string
	GetAvailablePaymentMethods() (GetAvailablePaymentMethodsResponse, error)
	CreateCheckoutPaymentSession(CreateCheckoutSessionPayload) (CreateCheckoutSessionResponse, error)
	CheckoutPaymentHandler(http.ResponseWriter, *http.Request) error
}
