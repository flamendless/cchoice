package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/client/components/layout"
	pb "cchoice/proto"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticate(string, string) (*pb.AuthLoginResponse, error)
	Authenticated(http.ResponseWriter, *http.Request) *common.HandlerRes
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
	return &common.HandlerRes{
		Component: layout.Base("Auth", components.AuthView()),
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

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if username == "" || password == "" {
		return &common.HandlerRes{
			Error:      errors.New("Incomplete form data"),
			StatusCode: http.StatusBadRequest,
		}
	}

	res, err := h.AuthService.Authenticate(username, password)
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	h.SM.Put(r.Context(), "tokenString", res.Token)

	return &common.HandlerRes{
		Component: layout.Base("Auth", components.AuthView()),
	}
}
