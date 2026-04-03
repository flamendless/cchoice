package server

import (
	"net/http"

	compcpoints "cchoice/cmd/web/components/cpoints"
	"cchoice/internal/logs"
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
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr == "" {
		redirectHX(w, r, utils.URLWithError("/customer", "Log in first"))
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		redirectHX(w, r, utils.URLWithError(page, "Token is required"))
		return
	}

	if err := s.services.cpoint.RedeemWithToken(ctx, token, customerIDStr); err != nil {
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
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr == "" {
		redirectHX(w, r, utils.URLWithError("/customer", "Log in first"))
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	code := r.PostFormValue("code")
	if code == "" {
		redirectHX(w, r, utils.URLWithError(page, "Code is required"))
		return
	}

	if err := s.services.cpoint.RedeemCpoint(ctx, customerIDStr, code); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/cpoints", "C-Points redeemed successfully!"))
}
