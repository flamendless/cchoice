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
	Authenticate(*pb.AuthenticateRequest) (*pb.AuthenticateResponse, error)
	Register(*pb.RegisterRequest) (*pb.RegisterResponse, error)
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

	_, err = h.AuthService.Register(&pb.RegisterRequest{
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

	return &common.HandlerRes{
		Component:  layout.Base("Log In", components.AuthView()),
		ReplaceURL: "/auth",
	}
}
