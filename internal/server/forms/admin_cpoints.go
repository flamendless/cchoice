package forms

type AdminCPointsGenerateForm struct {
	CustomerID  string `form:"customer-id" validate:"required"`
	Value       int64  `form:"value" validate:"required,gt=0"`
	ExpiresAt   string `form:"expires-at"`
	ProductSkus string `form:"product-skus"`
}

type AdminCPointsCodeQuery struct {
	Code       string `form:"code" validate:"required"`
	Redemption string `form:"redemption"`
}

type AdminCPointsQRQuery struct {
	Code string `form:"code" validate:"required"`
}
