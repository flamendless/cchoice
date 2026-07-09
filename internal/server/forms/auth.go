package forms

type ForgotPasswordQuery struct {
	Type string `form:"type" validate:"omitempty,user_type"`
}

type ForgotPasswordForm struct {
	Email    string `form:"email" validate:"required,ph_email"`
	UserType string `form:"user_type" validate:"required,user_type"`
}

type ResetPasswordPageQuery struct {
	Token string `form:"token" validate:"required"`
}

type ResetPasswordForm struct {
	Token           string `form:"token" validate:"required"`
	NewPassword     string `form:"new_password" validate:"required,ph_password"`
	ConfirmPassword string `form:"confirm_password" validate:"required,eqfield=NewPassword"`
}
