package server

import (
	"net/http"

	compauth "cchoice/cmd/web/components/auth"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddAuthHandlers(s *Server, r chi.Router) {
	r.Get("/auth/forgot-password", s.forgotPasswordPageHandler)
	r.Post("/auth/forgot-password", s.forgotPasswordHandler)
	r.Get("/auth/reset-password", s.resetPasswordPageHandler)
	r.Post("/auth/reset-password", s.resetPasswordHandler)
}

func (s *Server) forgotPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Forgot Password Page Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	userTypeStr := r.URL.Query().Get("type")
	userType := enums.ParseUserTypeToEnum(userTypeStr)
	if !userType.IsValid() && userTypeStr != "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	if err := compauth.ForgotPasswordPage(userType).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Forgot Password Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	userTypeStr := r.PostFormValue("user_type")
	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": err.Error(),
				"type":  userTypeStr,
			},
		))
		return
	}

	email := r.PostFormValue("email")
	if !constants.ReEmail.MatchString(email) {
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": errs.ErrInvalidParams.Error(),
				"type":  userTypeStr,
			},
		))
		return
	}

	userType := enums.ParseUserTypeToEnum(userTypeStr)
	if !userType.IsValid() {
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": errs.ErrInvalidParams.Error(),
				"type":  userTypeStr,
			},
		))
		return
	}

	if err := s.services.passwordReset.RequestReset(ctx, email, userType); err != nil {
		if err == errs.ErrPasswordResetRateLimited {
			redirectHX(w, r, utils.URLWithParams(
				page,
				map[string]string{
					"error": errs.ErrPasswordResetRateLimited.Error(),
					"type":  userTypeStr,
				},
			))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": err.Error(),
				"type":  userTypeStr,
			},
		))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Check your e-mail"))
	redirectHX(w, r, utils.URLWithParams(
		page,
		map[string]string{
			"success": "Check your e-mail",
			"type":    userTypeStr,
		},
	))
}

func (s *Server) resetPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Reset Password Page Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	token := r.URL.Query().Get("token")
	if token == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	resetCtx, err := s.services.passwordReset.VerifyToken(ctx, token)
	if err != nil {
		if err == errs.ErrInvalidResetToken {
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := compauth.ResetPasswordPage(resetCtx.UserType, token, resetCtx.Email).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Reset Password Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	token := r.PostFormValue("token")
	if token == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	newPassword := r.PostFormValue("new_password")
	confirmPassword := r.PostFormValue("confirm_password")
	if newPassword == "" || confirmPassword == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	if newPassword != confirmPassword {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	if !constants.RePassword.MatchString(newPassword) {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrInvalidParams.Error()))
		return
	}

	userType, err := s.services.passwordReset.ResetPassword(ctx, token, newPassword)
	if err != nil {
		if err == errs.ErrInvalidResetToken {
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	var url string
	switch userType {
	case enums.USER_TYPE_STAFF:
		url = "/admin"
	case enums.USER_TYPE_CUSTOMER:
		url = "/customer"
	}
	redirectHX(w, r, utils.URLWithSuccess(url, "Password reset successfully. Please login with your new password."))
}
