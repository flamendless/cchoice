package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/auth"
	"cchoice/internal/enums"
	"cchoice/internal/errs"

	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticated(enums.AudKind, http.ResponseWriter, *http.Request) (*common.HandlerRes, auth.ValidToken)
	Authenticate(*common.AuthAuthenticateRequest) (*common.AuthAuthenticateResponse, error)
	Register(*common.AuthRegisterRequest) (string, error)
	EnrollOTP(*common.AuthEnrollOTPRequest) (*common.AuthEnrollOTPResponse, error)
	FinishOTPEnrollment(*common.AuthFinishOTPEnrollment) error
	GetOTPCode(*common.AuthGetOTPCodeRequest) error
	GetOTPInfo(*common.AuthGetOTPInfoRequest) (*common.AuthGetOTPInfoResponse, error)
	ValidateToken(*common.AuthValidateTokenRequest) (*common.AuthValidateTokenResponse, error)
	ValidateOTP(*common.AuthValidateOTPRequest) error
}

type AuthHandler struct {
	Logger      *zap.Logger
	AuthService AuthService
	SM          *scs.SessionManager
}

type UserForRegistration struct {
	UserID   string
	EMail    string
	MobileNo string
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
			Component: components.Base(
				"Log In",
				components.CenterCard(components.AuthView()),
			),
			ReplaceURL: "/auth",
		}
	}
	return &common.HandlerRes{
		Component:  components.Base("Home"),
		ReplaceURL: "/home",
	}
}

func (h AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_PARSE_FORM,
			StatusCode: http.StatusBadRequest,
		}
	}

	res, err := h.AuthService.Authenticate(&common.AuthAuthenticateRequest{
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

	if res.NeedOTP {
		h.SM.Put(r.Context(), "needOTP", true)
		return &common.HandlerRes{
			Component: components.CenterCard(
				components.OTPView(false),
			),
			ReplaceURL: "/otp",
		}
	}

	return &common.HandlerRes{
		Component:  components.Base("Home"),
		ReplaceURL: "/home",
	}
}
