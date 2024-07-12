package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"errors"
	"net/http"
	"net/url"
)

func (h AuthHandler) OTPEnrollView(
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
			Component: components.Base(
				"OTP ENROLL",
				components.CenterCard(components.OTPView(true)),
			),
			RedirectTo: "/otp-enroll",
			ReplaceURL: "/otp-enroll",
		}
	}

	userForRegistration, ok := h.SM.Get(r.Context(), "userRegistration").(UserForRegistration)
	if !ok {
		return &common.HandlerRes{
			Error:      errors.New("Expired session. Register again"),
			StatusCode: http.StatusBadRequest,
		}
	}

	resInfo, err := h.AuthService.GetOTPInfo(&common.AuthGetOTPInfoRequest{
		UserID:    userForRegistration.UserID,
		OTPMethod: otpMethod.String(),
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	res, err := h.AuthService.EnrollOTP(&common.AuthEnrollOTPRequest{
		UserID:      userForRegistration.UserID,
		Issuer:      "cchoice",
		AccountName: resInfo.Recipient,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	if otpMethod == pb.OTPMethod_AUTHENTICATOR {
		imgSrc := serialize.PNGEncode(res.Image)
		return &common.HandlerRes{
			Component: components.OTPEnrollMethodAuthenticator(
				res.Secret,
				imgSrc,
				res.RecoveryCodes,
			),
			ReplaceURL: "/otp",
		}
	}

	if otpMethod == pb.OTPMethod_EMAIL || otpMethod == pb.OTPMethod_SMS {
		err = h.AuthService.GetOTPCode(&common.AuthGetOTPCodeRequest{
			UserID:    userForRegistration.UserID,
			OTPMethod: qOTPMethod,
		})
		if err != nil {
			return &common.HandlerRes{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}
		}

		return &common.HandlerRes{
			Component: components.OTPEnrollMethodSMSOrEMail(
				otpMethod.String(),
				resInfo.Recipient,
				res.RecoveryCodes,
			),
			ReplaceURL: "/otp",
		}
	}

	panic("should not be reached")
}

func (h AuthHandler) OTPEnrollFinish(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	err := r.ParseForm()
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_PARSE_FORM,
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

	err = h.AuthService.FinishOTPEnrollment(&common.AuthFinishOTPEnrollment{
		UserID:   userForRegistration.UserID,
		Passcode: passcode,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	_ = h.SM.PopString(r.Context(), "userRegistration")

	return &common.HandlerRes{
		Component: components.Base(
			"Log In",
			components.CenterCard(components.AuthView()),
		),
		ReplaceURL: "/auth",
	}
}
