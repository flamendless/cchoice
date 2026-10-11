package server

import (
	"net/http"
	"strconv"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

const adminTimeOffManagePage = "/admin/time-off"

func (s *Server) adminTimeOffManagePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Manage Page Handler]"
	ctx := r.Context()

	staffList, err := s.services.staff.ListActiveStaffForSelect(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if err := compadmin.AdminTimeOffManagePage(staffList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminTimeOffManageTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Manage Table Handler]"
	ctx := r.Context()

	timeOffs, err := s.services.staff.GetAllTimeOffs(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if err := compadmin.AdminTimeOffManageTable(timeOffs).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminTimeOffGrantCreditsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Grant Credits Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	var f forms.AdminStaffLeaveCreditGrantForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, httputil.ErrorMessage(err)))
		return
	}
	hours, err := strconv.ParseFloat(f.Hours, 64)
	if err != nil || hours <= 0 {
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, errs.ErrInvalidLeaveHours.Error()))
		return
	}
	staffID, err := httputil.RequireEncodedID(s.encoder, f.StaffID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, errs.ErrInvalidParams.Error()))
		return
	}
	if err := s.services.attendance.GrantLeaveCredits(ctx, staffID, currentStaffID, hours, f.Note); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, err.Error()))
		return
	}
	redirectHX(w, r, utils.URLWithSuccess(adminTimeOffManagePage, "Leave credits granted"))
}

func (s *Server) adminTimeOffApproveHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Approve Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	var p forms.AdminSuperuserTimeOffPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	timeOffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	if err := s.services.attendance.ApproveTimeOff(ctx, timeOffID, currentStaffID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, err.Error()))
		return
	}
	redirectHX(w, r, utils.URLWithSuccess(adminTimeOffManagePage, "Leave request approved"))
}

func (s *Server) adminTimeOffRejectModalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p forms.AdminSuperuserTimeOffPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	timeOffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	if err := compadmin.TimeOffRejectModal(timeOffID).Render(ctx, w); err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminTimeOffRejectHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Reject Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	var p forms.AdminSuperuserTimeOffPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	timeOffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, errs.ErrInvalidParams.Error()))
		return
	}

	var f forms.AdminTimeOffRejectForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, httputil.ErrorMessage(err)))
		return
	}
	if err := s.services.attendance.RejectTimeOff(ctx, timeOffID, currentStaffID, f.RejectReason); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, err.Error()))
		return
	}
	redirectHX(w, r, utils.URLWithSuccess(adminTimeOffManagePage, "Leave request rejected"))
}

func (s *Server) adminTimeOffCancelApprovalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Time Off Cancel Approval Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	var p forms.AdminSuperuserTimeOffPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	timeOffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	if err := s.services.attendance.CancelTimeOff(ctx, timeOffID, currentStaffID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffID), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminTimeOffManagePage, err.Error()))
		return
	}
	redirectHX(w, r, utils.URLWithSuccess(adminTimeOffManagePage, "Leave approval reverted to pending"))
}

// Legacy superuser routes delegate to the shared manage handlers.
func (s *Server) adminSuperuserTimeOffPageHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffManagePageHandler(w, r)
}

func (s *Server) adminSuperuserTimeOffTableHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffManageTableHandler(w, r)
}

func (s *Server) adminSuperuserTimeOffApproveHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffApproveHandler(w, r)
}

func (s *Server) adminSuperuserTimeOffRejectModalHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffRejectModalHandler(w, r)
}

func (s *Server) adminSuperuserTimeOffRejectHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffRejectHandler(w, r)
}

func (s *Server) adminSuperuserTimeOffCancelHandler(w http.ResponseWriter, r *http.Request) {
	s.adminTimeOffCancelApprovalHandler(w, r)
}
