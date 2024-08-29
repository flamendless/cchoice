package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/client/middlewares"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	pb "cchoice/proto"

	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type AuthService interface {
	pb.AuthServiceClient
}

type AuthHandler struct {
	Logger        *zap.Logger
	AuthService   AuthService
	SM            *scs.SessionManager
	Authenticated *middlewares.Authenticated
}

func NewAuthHandler(
	logger *zap.Logger,
	authService AuthService,
	sm *scs.SessionManager,
	authenticated *middlewares.Authenticated,
) AuthHandler {
	return AuthHandler{
		Logger:        logger,
		AuthService:   authService,
		SM:            sm,
		Authenticated: authenticated,
	}
}

func (h AuthHandler) AuthPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	_, err := h.Authenticated.Authenticated(w, r, enums.AUD_API)
	if err != nil {
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

	res, err := h.AuthService.Authenticate(
		r.Context(),
		&pb.AuthenticateRequest{
			Username: r.Form.Get("username"),
			Password: r.Form.Get("password"),
		},
	)
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	h.SM.Put(r.Context(), "authSession", common.AuthSession{
		Token:   res.Token,
		NeedOTP: res.NeedOtp,
	})

	return &common.HandlerRes{
		Component: components.Base(
			"OTP",
			components.CenterCard(
				components.OTPView(false),
			),
		),
		ReplaceURL: "/otp",
	}
}

func (h AuthHandler) Avatar(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	user, err := h.Authenticated.User(w, r, enums.AUD_API)
	if err != nil {
		return &common.HandlerRes{
			Component: components.UserAvatar(""),
		}
	}
	return &common.HandlerRes{
		Component: components.UserAvatar(user.FirstName),
	}
}
