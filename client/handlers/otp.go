package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"errors"
	"net/http"
	"net/url"
)

func (h AuthHandler) OTPView(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	needOTP := h.SM.GetBool(r.Context(), "needOTP")
	tokenString := h.SM.GetString(r.Context(), "tokenString")
	if !needOTP || tokenString == "" {
		return &common.HandlerRes{
			Error:      errs.ERR_EXPIRED_OTP_LOGIN_AGAIN,
			StatusCode: http.StatusBadRequest,
		}
	}

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
			Component: components.Base(
				"OTP",
				components.CenterCard(components.OTPView(false)),
			),
			RedirectTo: "/otp",
			ReplaceURL: "/otp",
		}
	}

	resValidateToken, err := h.AuthService.ValidateToken(&common.AuthValidateTokenRequest{
		Token: tokenString,
		AUD:   "API",
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_NO_AUTH,
			StatusCode: http.StatusUnauthorized,
		}
	}

	resInfo, err := h.AuthService.GetOTPInfo(&common.AuthGetOTPInfoRequest{
		UserID:    resValidateToken.UserID,
		OTPMethod: otpMethod.String(),
	})

	if otpMethod == pb.OTPMethod_AUTHENTICATOR {
		return &common.HandlerRes{
			Component:  components.OTPMethodAuthenticator(),
			ReplaceURL: "/otp",
		}
	}

	if otpMethod == pb.OTPMethod_EMAIL || otpMethod == pb.OTPMethod_SMS {
		err = h.AuthService.GetOTPCode(&common.AuthGetOTPCodeRequest{
			UserID:    resValidateToken.UserID,
			OTPMethod: qOTPMethod,
		})
		if err != nil {
			return &common.HandlerRes{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		return &common.HandlerRes{
			Component: components.OTPMethodSMSOrEMail(
				otpMethod.String(),
				resInfo.Recipient,
			),
			ReplaceURL: "/otp",
		}
	}

	panic("should not be reached")
}

func (h AuthHandler) OTPValidate(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	resAuth, validToken := h.AuthService.Authenticated(enums.AUD_API, w, r)
	if resAuth != nil {
		return resAuth
	}

	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_PARSE_FORM,
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

	err = h.AuthService.ValidateOTP(&common.AuthValidateOTPRequest{
		UserID:   validToken.UserID,
		Passcode: passcode,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	_ : h.SM.PopBool(r.Context(), "needOTP")

	return &common.HandlerRes{
		Component:  components.Base("Home"),
		ReplaceURL: "/home",
	}
}
