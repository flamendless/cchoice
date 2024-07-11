package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/enums"

	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticated(enums.AudKind, http.ResponseWriter, *http.Request) *common.HandlerRes
	Authenticate(*common.AuthAuthenticateRequest) (string, error)
	Register(*common.AuthRegisterRequest) (string, error)
	EnrollOTP(*common.AuthEnrollOTPRequest) (*common.AuthEnrollOTPResponse, error)
	FinishOTPEnrollment(*common.AuthFinishOTPEnrollment) error
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
			Component:  components.Base(
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
		Component:  components.Base("Home"),
		ReplaceURL: "/home",
	}
}

