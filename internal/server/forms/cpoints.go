package forms

type CPointsClaimQuery struct {
	Token string `form:"token" validate:"required"`
}

type CPointsRedeemForm struct {
	Code string `form:"code" validate:"required"`
}
