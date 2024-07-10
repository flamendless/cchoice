package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/internal/enums"

	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"errors"
	"net/http"
	"net/url"

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

type UserForRegistration struct {
	UserID      string
	EMail string
	MobileNo    string
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

	h.SM.Put(
		r.Context(),
		"userRegistration",
		&UserForRegistration{
			UserID:      userID,
			EMail: r.Form.Get("email"),
			MobileNo:    r.Form.Get("mobile_no"),
		},
	)

	return &common.HandlerRes{
		Component: layout.Base(
			"OTP",
			components.OTPView(pb.OTPMethod_UNDEFINED),
		),
		RedirectTo: "/otp",
		ReplaceURL: "/otp",
	}
}

func (h AuthHandler) OTPView(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	qOTPMethod := q.Get("otp_method")
	otpMethod := enums.StringToPBEnum(
		qOTPMethod,
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)
	if otpMethod == pb.OTPMethod_UNDEFINED {
		return &common.HandlerRes{
			Component: layout.Base(
				"OTP",
				components.OTPView(otpMethod),
			),
			ReplaceURL: "/otp",
		}
	}

	userForRegistration, ok := h.SM.Get(r.Context(), "userRegistration").(UserForRegistration)
	if !ok {
		return &common.HandlerRes{
			Error:      errors.New("Expired session. Register again"),
			StatusCode: http.StatusBadRequest,
		}
	}

	if otpMethod == pb.OTPMethod_AUTHENTICATOR {
		res, err := h.AuthService.EnrollOTP(&common.AuthEnrollOTPRequest{
			UserID:      userForRegistration.UserID,
			Issuer:      "cchoice",
			AccountName: userForRegistration.EMail,
		})
		if err != nil {
			return &common.HandlerRes{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		imgSrc := serialize.PNGEncode(res.Image)
		return &common.HandlerRes{
			Component: components.OTPMethodAuthenticate(
				res.Secret,
				imgSrc,
				res.RecoveryCodes,
			),
			ReplaceURL: "/otp",
		}
	}

	if otpMethod == pb.OTPMethod_EMAIL || otpMethod == pb.OTPMethod_SMS {
		res, err := h.AuthService.EnrollOTP(&common.AuthEnrollOTPRequest{
			UserID:      userForRegistration.UserID,
			Issuer:      "cchoice",
			AccountName: userForRegistration.EMail,
		})
		if err != nil {
			return &common.HandlerRes{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		err = h.AuthService.GetOTPCode(&common.AuthGetOTPCodeRequest{
			UserID: userForRegistration.UserID,
			Method: qOTPMethod,
		})
		if err != nil {
			return &common.HandlerRes{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		if otpMethod == pb.OTPMethod_SMS {
			return &common.HandlerRes{
				Component: components.OTPMethodSMSOrEMail(
					"SMS",
					userForRegistration.MobileNo,
					res.RecoveryCodes,
				),
				ReplaceURL: "/otp",
			}
		} else if otpMethod == pb.OTPMethod_EMAIL {
			return &common.HandlerRes{
				Component: components.OTPMethodSMSOrEMail(
					"E-Mail",
					userForRegistration.EMail,
					res.RecoveryCodes,
				),
				ReplaceURL: "/otp",
			}
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

	userForRegistration, ok := h.SM.Get(r.Context(), "userRegistration").(UserForRegistration)
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
		UserID:   userForRegistration.UserID,
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
