package server

import (
	"net/http"

	compauth "cchoice/cmd/web/components/auth"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddAuthHandlers(s *Server, r chi.Router) {
	r.Get("/auth/forgot-password", s.forgotPasswordPageHandler)
	r.Get("/auth/reset-password", s.resetPasswordPageHandler)
	r.Group(func(r chi.Router) {
		r.Use(s.rateLimiter.Middleware)
		r.Post("/auth/forgot-password", s.forgotPasswordHandler)
		r.Post("/auth/reset-password", s.resetPasswordHandler)
	})
}

func (s *Server) forgotPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Forgot Password Page Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	var req forms.ForgotPasswordQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	userType := enums.ParseUserTypeToEnum(req.Type)
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

	var req forms.ForgotPasswordForm
	if err := httputil.BindPostForm(r, &req); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": httputil.ErrorMessage(err),
				"type":  req.UserType,
			},
		))
		return
	}

	userType := enums.ParseUserTypeToEnum(req.UserType)
	if err := s.services.passwordReset.RequestReset(ctx, req.Email, userType); err != nil {
		if err == errs.ErrPasswordResetRateLimited {
			redirectHX(w, r, utils.URLWithParams(
				page,
				map[string]string{
					"error": errs.ErrPasswordResetRateLimited.Error(),
					"type":  req.UserType,
				},
			))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithParams(
			page,
			map[string]string{
				"error": err.Error(),
				"type":  req.UserType,
			},
		))
		return
	}

	redirectHX(w, r, utils.URLWithParams(
		page,
		map[string]string{
			"success": "Check your e-mail",
			"type":    req.UserType,
		},
	))
}

func (s *Server) resetPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Reset Password Page Handler]"
	const page = "/auth/forgot-password"
	ctx := r.Context()

	var req forms.ResetPasswordPageQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	resetCtx, err := s.services.passwordReset.VerifyToken(ctx, req.Token)
	if err != nil {
		if err == errs.ErrInvalidResetToken {
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if err := compauth.ResetPasswordPage(resetCtx.UserType, req.Token, resetCtx.Email).Render(ctx, w); err != nil {
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

	var req forms.ResetPasswordForm
	if err := httputil.BindPostForm(r, &req); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	userType, err := s.services.passwordReset.ResetPassword(ctx, req.Token, req.NewPassword)
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
