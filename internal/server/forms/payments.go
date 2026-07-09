package forms

type PaymentRedirectQuery struct {
	PaymentRef string `form:"payment_ref" validate:"required"`
	Token      string `form:"token"`
}
