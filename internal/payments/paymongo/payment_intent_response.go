package paymongo

type GetPaymentIntentResponse struct {
	Data PaymentIntentData `json:"data"`
}

type PaymentIntentData struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Attributes PaymentIntentAttributes `json:"attributes"`
}
