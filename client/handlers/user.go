package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type UserService interface {
	pb.UserServiceClient
}

type UserHandler struct {
	Logger      *zap.Logger
	UserService UserService
	SM          *scs.SessionManager
}

func NewUserHandler(
	logger *zap.Logger,
	userService UserService,
	sm *scs.SessionManager,
) UserHandler {
	return UserHandler{
		Logger:      logger,
		UserService: userService,
		SM:          sm,
	}
}

func (h UserHandler) RegisterPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	return &common.HandlerRes{
		Component: components.Base(
			"Register",
			components.CenterCard(components.RegisterView()),
		),
	}
}

func (h UserHandler) Register(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_PARSE_FORM,
			StatusCode: http.StatusBadRequest,
		}
	}

	resRegister, err := h.UserService.Register(
		r.Context(),
		&pb.RegisterRequest{
			FirstName:       r.Form.Get("first_name"),
			MiddleName:      r.Form.Get("middle_name"),
			LastName:        r.Form.Get("last_name"),
			Email:           r.Form.Get("email"),
			Password:        r.Form.Get("password"),
			ConfirmPassword: r.Form.Get("confirm_password"),
			MobileNo:        r.Form.Get("mobile_no"),
		},
	)
	if err != nil {
		return &common.HandlerRes{Error: err}
	}

	h.SM.Put(r.Context(), "EncUserID", resRegister.UserId)

	return &common.HandlerRes{
		Component: components.Base(
			"OTP ENROLL",
			components.CenterCard(components.OTPView(true)),
		),
		RedirectTo: "/otp-enroll",
		ReplaceURL: "/otp-enroll",
	}
}
