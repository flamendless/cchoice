package server

import (
	"context"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/encode"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) adminStaffMemoAcceptHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Memo Accept Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	var p forms.AdminStaffMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), httputil.ErrorMessage(err)))
		return
	}
	memoID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
		return
	}

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

	var p forms.AdminStaffMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), httputil.ErrorMessage(err)))
		return
	}
	memoID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
		return
	}
	if err := compadmin.StaffMemoRejectModal(memoID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
	}
}

func (s *Server) adminStaffMemoRejectHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Memo Reject Handler]"
	ctx := r.Context()

	var p forms.AdminStaffMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), httputil.ErrorMessage(err)))
		return
	}
	memoID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), err.Error()))
		return
	}
	var f forms.AdminStaffMemoRejectForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(s.staffHomeRedirect(ctx), httputil.ErrorMessage(err)))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	reason := f.RejectReason

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
