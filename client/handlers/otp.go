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

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

type OTPService interface {
	pb.OTPServiceClient
}

type OTPHandler struct {
	Logger      *zap.Logger
	OTPService  OTPService
	AuthService AuthService
	SM          *scs.SessionManager
}

func NewOTPHandler(
	logger *zap.Logger,
	otpService OTPService,
	sm *scs.SessionManager,
) OTPHandler {
	return OTPHandler{
		Logger:      logger,
		OTPService:  otpService,
		SM:          sm,
	}
}

func (h OTPHandler) OTPView(
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

	// resValidateToken, err := h.OTPService.ValidateToken(&common.AuthValidateTokenRequest{
	// 	Token: tokenString,
	// 	AUD:   "API",
	// })
	// if err != nil {
	// 	return &common.HandlerRes{
	// 		Error:      errs.ERR_NO_AUTH,
	// 		StatusCode: http.StatusUnauthorized,
	// 	}
	// }
	//
	// resInfo, err := h.OTPService.GetOTPInfo(&common.AuthGetOTPInfoRequest{
	// 	UserID:    resValidateToken.UserID,
	// 	OTPMethod: otpMethod.String(),
	// })
	//
	// if otpMethod == pb.OTPMethod_AUTHENTICATOR {
	// 	return &common.HandlerRes{
	// 		Component:  components.OTPMethodAuthenticator(),
	// 		ReplaceURL: "/otp",
	// 	}
	// }
	//
	// if otpMethod == pb.OTPMethod_EMAIL || otpMethod == pb.OTPMethod_SMS {
	// 	err = h.OTPService.GetOTPCode(&common.AuthGetOTPCodeRequest{
	// 		UserID:    resValidateToken.UserID,
	// 		OTPMethod: qOTPMethod,
	// 	})
	// 	if err != nil {
	// 		return &common.HandlerRes{
	// 			Error:      err,
	// 			StatusCode: http.StatusInternalServerError,
	// 		}
	// 	}
	//
	// 	return &common.HandlerRes{
	// 		Component: components.OTPMethodSMSOrEMail(
	// 			otpMethod.String(),
	// 			resInfo.Recipient,
	// 		),
	// 		ReplaceURL: "/otp",
	// 	}
	// }

	panic("should not be reached")
}

func (h OTPHandler) OTPValidate(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	// validToken, err := h.AuthService.Authenticated(enums.AUD_API, w, r)
	// if err != nil {
	// 	return &common.HandlerRes{
	// 		Error: err,
	// 	}
	// }

	// err = r.ParseForm()
	// if err != nil {
	// 	return &common.HandlerRes{
	// 		Error:      errs.ERR_PARSE_FORM,
	// 		StatusCode: http.StatusBadRequest,
	// 	}
	// }

	passcode := r.Form.Get("otp")
	if passcode == "" {
		return &common.HandlerRes{
			Error:      errors.New("Invalid OTP"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// err = h.OTPService.ValidateOTP(&common.AuthValidateOTPRequest{
	// 	UserID:   validToken.UserID,
	// 	Passcode: passcode,
	// })
	// if err != nil {
	// 	return &common.HandlerRes{
	// 		Error:      err,
	// 		StatusCode: http.StatusBadRequest,
	// 	}
	// }

	h.SM.PopBool(r.Context(), "needOTP")

	return &common.HandlerRes{
		Component:  components.Base("Home"),
		ReplaceURL: "/home",
	}
}

func (h OTPHandler) OTPEnrollView(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	encUserID := h.SM.GetString(r.Context(), "EncUserID")
	if encUserID == "" {
		return &common.HandlerRes{
			Error: errs.ERR_EXPIRED_REGISTRATION,
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
				"OTP ENROLL",
				components.CenterCard(components.OTPView(true)),
			),
			RedirectTo: "/otp-enroll",
			ReplaceURL: "/otp-enroll",
		}
	}

	resInfo, err := h.OTPService.GetOTPInfo(
		r.Context(),
		&pb.GetOTPInfoRequest{
			UserId:    encUserID,
			OtpMethod: otpMethod.String(),
		},
	)
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	res, err := h.OTPService.EnrollOTP(
		r.Context(),
		&pb.EnrollOTPRequest{
			UserId:      encUserID,
			Issuer:      "cchoice",
			AccountName: resInfo.Recipient,
		},
	)
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
		_, err = h.OTPService.GenerateOTPCode(
			r.Context(),
			&pb.GenerateOTPCodeRequest{
				UserId: encUserID,
				Method: otpMethod,
			},
		)
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

func (h OTPHandler) OTPEnrollFinish(
	w http.ResponseWriter,
	r *http.Request,
) *common.HandlerRes {
	encUserID := h.SM.PopString(r.Context(), "EncUserID")
	if encUserID == "" {
		return &common.HandlerRes{
			Error: errs.ERR_EXPIRED_REGISTRATION,
		}
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
			Error:      errs.ERR_INVALID_OTP,
			StatusCode: http.StatusBadRequest,
		}
	}

	_, err = h.OTPService.FinishOTPEnrollment(
		r.Context(),
		&pb.FinishOTPEnrollmentRequest{
			UserId:   encUserID,
			Passcode: passcode,
		},
	)
	if err != nil {
		return &common.HandlerRes{
			Error:      err,
			StatusCode: http.StatusBadRequest,
		}
	}

	return &common.HandlerRes{
		Component: components.Base(
			"Log In",
			components.CenterCard(components.AuthView()),
		),
		ReplaceURL: "/auth",
	}
}
