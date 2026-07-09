package forms

type AdminLoginForm struct {
	Email       string `form:"email" validate:"required,ph_email"`
	Password    string `form:"password" validate:"required,ph_password"`
	LocationLat string `form:"location_lat"`
	LocationLng string `form:"location_lng"`
}
