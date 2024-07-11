package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"errors"
	"net/http"
)

func (h AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	return &common.HandlerRes{
		Component: components.Base(
			"Register",
			components.CenterCard(components.RegisterView()),
		),
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
		Component: components.Base(
			"OTP ENROLL",
			components.CenterCard(components.OTPView(true)),
		),
		RedirectTo: "/otp-enroll",
		ReplaceURL: "/otp-enroll",
	}
}

