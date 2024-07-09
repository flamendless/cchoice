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

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticate(*pb.AuthenticateRequest) (*pb.AuthenticateResponse, error)
	Register(*pb.RegisterRequest) (*pb.RegisterResponse, error)
	Authenticated(http.ResponseWriter, *http.Request) *common.HandlerRes
	EnrollOTP(*pb.EnrollOTPRequest) (*pb.EnrollOTPResponse, error)
	ValidateInitialOTP(*pb.ValidateInitialOTPRequest) (*pb.ValidateInitialOTPResponse, error)
	GetOTPCode(*pb.GetOTPCodeRequest) (*pb.GetOTPCodeResponse, error)
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

	res, err := h.AuthService.Authenticate(&pb.AuthenticateRequest{
		Username: r.Form.Get("username"),
		Password: r.Form.Get("password"),
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	h.SM.Put(r.Context(), "tokenString", res.Token)

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

	resRegister, err := h.AuthService.Register(&pb.RegisterRequest{
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

	res, err := h.AuthService.EnrollOTP(&pb.EnrollOTPRequest{
		UserId:      resRegister.UserId,
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
		resRegister.UserId,
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

	_, err = h.AuthService.GetOTPCode(
		&pb.GetOTPCodeRequest{
			Method: enums.StringToPBEnum(
				r.Form.Get("method"),
				pb.OTPMethod_OTPMethod_value,
				pb.OTPMethod_UNDEFINED,
			),
		},
	)
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

	_, err = h.AuthService.ValidateInitialOTP(&pb.ValidateInitialOTPRequest{
		UserId:   userID,
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
