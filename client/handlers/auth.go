package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/internal/serialize"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticated(http.ResponseWriter, *http.Request) *common.HandlerRes
	Authenticate(*common.AuthAuthenticateRequest) (string, error)
	Register(*common.AuthRegisterRequest) (string, error)
	EnrollOTP(*common.AuthEnrollOTPRequest) (*common.AuthEnrollOTPResponse, error)
	ValidateInitialOTP(*common.AuthValidateInitialOTP) error
	GetOTPCode(*common.AuthGetOTPCodeRequest) error
}

type AuthHandler struct {
	Logger      *zap.Logger
	AuthService AuthService
	SM          *scs.SessionManager
}

func NewAuthHandler(
	logger *zap.Logger,
	authService AuthService,
	sm *scs.SessionManager,
) AuthHandler {
	return AuthHandler{
		Logger:      logger,
		AuthService: authService,
		SM:          sm,
	}
}

func (h AuthHandler) AuthPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	tokenString := h.SM.GetString(r.Context(), "tokenString")
	if tokenString == "" {
		return &common.HandlerRes{
			Component:  layout.Base("Log In", components.AuthView()),
			ReplaceURL: "/auth",
		}
	}
	return &common.HandlerRes{
		Component:  layout.Base("Home"),
		ReplaceURL: "/home",
	}
}

func (h AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	return &common.HandlerRes{
		Component: layout.Base("Register", components.RegisterView()),
	}
}

func (h AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errors.New("Failed to parse form"),
			StatusCode: http.StatusBadRequest,
		}
	}

	tokenString, err := h.AuthService.Authenticate(&common.AuthAuthenticateRequest{
		Username: r.Form.Get("username"),
		Password: r.Form.Get("password"),
	})
	if err != nil || tokenString == "" {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	h.SM.Put(r.Context(), "tokenString", tokenString)

	return &common.HandlerRes{
		Component:  layout.Base("Home"),
		ReplaceURL: "/home",
	}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errors.New("Failed to parse form"),
			StatusCode: http.StatusBadRequest,
		}
	}

	userID, err := h.AuthService.Register(&common.AuthRegisterRequest{
		FirstName:       r.Form.Get("first_name"),
		MiddleName:      r.Form.Get("middle_name"),
		LastName:        r.Form.Get("last_name"),
		Email:           r.Form.Get("email"),
		Password:        r.Form.Get("password"),
		ConfirmPassword: r.Form.Get("confirm_password"),
		MobileNo:        r.Form.Get("mobile_no"),
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	res, err := h.AuthService.EnrollOTP(&common.AuthEnrollOTPRequest{
		UserID:      userID,
		Issuer:      "cchoice",
		AccountName: r.Form.Get("email"),
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	h.SM.Put(
		r.Context(),
		"userRegistration",
		userID,
	)

	imgSrc := serialize.PNGEncode(res.Image)
	return &common.HandlerRes{
		Component: layout.Base(
			"OTP Setup",
			components.OTPSetupView(res.Secret, imgSrc, res.RecoveryCodes),
		),
		ReplaceURL: "/otp-setup",
	}
}

func (h AuthHandler) GetOTPCode(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errors.New("Failed to parse form"),
			StatusCode: http.StatusBadRequest,
		}
	}

	err = h.AuthService.GetOTPCode(&common.AuthGetOTPCodeRequest{
		Method: r.Form.Get("method"),
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	return nil
}

func (h AuthHandler) ValidateInitialOTP(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errors.New("Failed to parse form"),
			StatusCode: http.StatusBadRequest,
		}
	}

	userID, ok := h.SM.Get(r.Context(), "userRegistration").(string)
	if !ok {
		return &common.HandlerRes{
			Error:      errors.New("Expired session. Register again"),
			StatusCode: http.StatusBadRequest,
		}
	}

	passcode := r.Form.Get("otp")
	if passcode == "" {
		return &common.HandlerRes{
			Error:      errors.New("Invalid OTP"),
			StatusCode: http.StatusBadRequest,
		}
	}

	err = h.AuthService.ValidateInitialOTP(&common.AuthValidateInitialOTP{
		UserID:   userID,
		Passcode: passcode,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	return &common.HandlerRes{
		Component:  layout.Base("Log In", components.AuthView()),
		ReplaceURL: "/auth",
	}
}
