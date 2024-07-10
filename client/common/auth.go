package common

type AuthAuthenticateRequest struct {
	Username string
	Password string
}

type AuthRegisterRequest struct {
	FirstName       string
	MiddleName      string
	LastName        string
	Email           string
	Password        string
	ConfirmPassword string
	MobileNo        string
}

type AuthEnrollOTPRequest struct {
	UserID      string
	Issuer      string
	AccountName string
}

type AuthEnrollOTPResponse struct {
	Secret        string
	RecoveryCodes string
	Image         []byte
}

type AuthValidateInitialOTP struct {
	UserID   string
	Passcode string
}

type AuthGetOTPCodeRequest struct {
	UserID string
	Method string
}
