package server

import (
	"net/http"

	compcpoints "cchoice/cmd/web/components/cpoints"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddCPointsHandlers(s *Server, r chi.Router) {
	r.With(s.requireCustomerAuth).Get("/cpoints", s.cpointsHomeHandler)
	r.With(s.requireCustomerAuth).Get("/cpoints/total", s.cpointsTotalHandler)
	r.With(s.requireCustomerAuth).Get("/cpoints/claim", s.cpointsClaimHandler)
	r.With(s.requireCustomerAuth).Get("/cpoints/redeem", s.cpointsRedeemPageHandler)
	r.With(s.requireCustomerAuth).Post("/cpoints/redeem", s.cpointsRedeemHandler)
}

func (s *Server) cpointsHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[C-Points Home Handler]"
	ctx := r.Context()

	if err := compcpoints.CPointsHomePage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) cpointsTotalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[C-Points Total Handler]"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cpoints, err := s.services.cpoint.GetRedeemedCpointsByCustomerID(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Failed to load cpoints", http.StatusInternalServerError)
		return
	}

	if err := compcpoints.CPointsTotal(cpoints.Total).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) cpointsClaimHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[C-Points Claim Handler]"
	const page = "/cpoints"
	const loginPage = "/customer"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr == "" {
		redirectHX(w, r, utils.URLWithError(loginPage, errs.ErrLogInFirst.Error()))
		return
	}

	var req forms.CPointsClaimQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	if err := s.services.cpoint.RedeemWithToken(ctx, req.Token, customerIDStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "C-Points redeemed successfully!"))
}

func (s *Server) cpointsRedeemPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[C-Points Redeem Page Handler]"
	ctx := r.Context()

	if err := compcpoints.CPointsRedeemPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) cpointsRedeemHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[C-Points Redeem Handler]"
	const page = "/cpoints/redeem"
	const successPage = "/cpoints"
	const loginPage = "/customer"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr == "" {
		redirectHX(w, r, utils.URLWithError(loginPage, errs.ErrLogInFirst.Error()))
		return
	}

	var req forms.CPointsRedeemForm
	if err := httputil.BindPostForm(r, &req); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	if err := s.services.cpoint.RedeemCpoint(ctx, customerIDStr, req.Code); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(successPage, "C-Points redeemed successfully!"))
}
