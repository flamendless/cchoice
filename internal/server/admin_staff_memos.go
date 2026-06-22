package server

import (
	"context"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminStaffMemoAcceptHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Memo Accept Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	memoID := chi.URLParam(r, "id")

	if err := s.services.memo.AcceptMemo(ctx, staffID, memoID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(s.staffHomeRedirect(ctx), "Memo accepted"))
}

func (s *Server) adminStaffMemoRejectModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Memo Reject Modal Handler]"
	ctx := r.Context()

	memoID := chi.URLParam(r, "id")
	if err := compadmin.StaffMemoRejectModal(memoID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
	}
}

func (s *Server) adminStaffMemoRejectHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Memo Reject Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), "Failed to parse form"))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	memoID := chi.URLParam(r, "id")
	reason := r.FormValue("reject_reason")

	if err := s.services.memo.RejectMemo(ctx, staffID, memoID, reason); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(s.staffHomeRedirect(ctx), "Memo rejected"))
}

func (s *Server) staffHomeRedirect(ctx context.Context) string {
	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staffID := s.encoder.Decode(staffIDStr)
	if staffID == encode.INVALID {
		return "/admin"
	}

	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		return "/admin/staff"
	}

	switch staff.UserType {
	case "SUPERUSER":
		return "/admin/superuser"
	default:
		return "/admin/staff"
	}
}
