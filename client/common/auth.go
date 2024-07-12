package common

type AuthAuthenticateRequest struct {
	Username string
	Password string
}

type AuthAuthenticateResponse struct {
	Token   string
	NeedOTP bool
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

type AuthFinishOTPEnrollment struct {
	UserID   string
	Passcode string
}

type AuthGetOTPCodeRequest struct {
	UserID    string
	OTPMethod string
}

type AuthGetOTPInfoRequest struct {
	UserID    string
	OTPMethod string
}

type AuthGetOTPInfoResponse struct {
	Recipient string
}

type AuthValidateTokenRequest struct {
	Token string
	AUD   string
}

type AuthValidateTokenResponse struct {
	UserID string
}

type AuthValidateOTPRequest struct {
	UserID   string
	Passcode string
}
