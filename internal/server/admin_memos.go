package server

import (
	"net/http"
	"strings"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminMemosListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos List Page Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	if err := compadmin.AdminMemosListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminMemosListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos List Table Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	serviceMemos, err := s.services.memo.GetAllForAdmin(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	isSuperuser := false
	if staff, staffErr := s.services.staff.GetCurrentStaff(ctx, currentStaffID); staffErr == nil {
		isSuperuser = staff.UserType == enums.STAFF_USER_TYPE_SUPERUSER.String()
	}

	memos := make([]models.AdminMemoListItem, 0, len(serviceMemos))
	for _, m := range serviceMemos {
		memos = append(memos, models.AdminMemoListItem{
			ID:            s.encoder.Encode(m.ID),
			Title:         m.Title,
			Message:       m.Message,
			FileURL:       m.FileURL,
			Status:        m.Status,
			StartDate:     m.StartDate,
			EndDate:       m.EndDate,
			CreatedByID:   s.encoder.Encode(m.CreatedBy),
			CreatedByName: m.CreatedByName,
			CreatedAt:     m.CreatedAt.Format(constants.DateTimeLayoutISO),
			EmailsSentAt:  m.EmailsSentAt,
		})
	}

	if err := compadmin.AdminMemosListTable(memos, currentStaffID, isSuperuser).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminMemosStaffRowsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Staff Rows Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var p forms.AdminMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	rows, err := s.services.memo.GetRecipientsWithActions(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	recipientRows := make([]models.AdminMemoRecipientRow, 0, len(rows))
	for _, row := range rows {
		recipientRows = append(recipientRows, models.AdminMemoRecipientRow{
			StaffID:      row.StaffID,
			StaffName:    row.StaffName,
			Email:        row.Email,
			Position:     row.Position,
			UserType:     row.UserType,
			ActionStatus: row.ActionStatus,
			RejectReason: row.RejectReason,
			AcceptedAt:   row.AcceptedAt,
			RejectedAt:   row.RejectedAt,
		})
	}

	if err := compadmin.MemoStaffRows(idStr, recipientRows, currentStaffID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminMemosCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Create Page Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	staffList, err := s.services.staff.GetAllForMemo(ctx, 100)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := compadmin.MemoCreateModal(staffList, nil, currentStaffID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminMemosEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Edit Page Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var p forms.AdminMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoGetFailed.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoGetFailed.Error()))
		return
	}
	memo, err := s.services.memo.GetMemoByID(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoGetFailed.Error()))
		return
	}
	if memo == nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoNotFound.Error()))
		return
	}

	recipientIDs, err := s.services.memo.GetRecipientStaffIDs(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	staffList, err := s.services.staff.GetAllForMemo(ctx, maxStaffListSize)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	memoItem := models.AdminMemoListItem{
		ID:           idStr,
		Title:        memo.Title,
		Message:      memo.Message,
		FileURL:      memo.FileURL,
		Status:       memo.Status,
		StartDate:    memo.StartDate,
		EndDate:      memo.EndDate,
		RecipientIDs: recipientIDs,
	}

	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := compadmin.MemoEditModal(memoItem, staffList, currentStaffID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminMemosCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Create Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var f forms.AdminMemoForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	title := f.Title
	message := f.Message
	fileURL := strings.TrimSpace(f.FileURL)
	statusStr := f.Status
	startDateStr := f.StartDate
	endDateStr := f.EndDate
	staffIDs := f.StaffIDs

	if title == "" || message == "" || statusStr == "" || startDateStr == "" || endDateStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoAllFieldsRequired.Error()))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	if err := services.ValidateMemoDates(startDate, endDate); err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	status := enums.ParseMemoStatusToEnum(statusStr)
	if status == enums.MEMO_STATUS_UNDEFINED || status == enums.MEMO_STATUS_DELETED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoInvalidStatus.Error()))
		return
	}

	if _, err := s.services.memo.CreateMemo(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		title,
		message,
		fileURL,
		status,
		startDate,
		endDate,
		staffIDs,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Memo created successfully"))
}

func (s *Server) adminMemosUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Update Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var p forms.AdminMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoAllFieldsRequired.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoAllFieldsRequired.Error()))
		return
	}
	var f forms.AdminMemoForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	title := f.Title
	message := f.Message
	fileURL := strings.TrimSpace(f.FileURL)
	statusStr := f.Status
	startDateStr := f.StartDate
	endDateStr := f.EndDate
	staffIDs := f.StaffIDs

	if idStr == "" || title == "" || message == "" || statusStr == "" || startDateStr == "" || endDateStr == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoAllFieldsRequired.Error()))
		return
	}

	status := enums.ParseMemoStatusToEnum(statusStr)
	if status == enums.MEMO_STATUS_DELETED {
		s.adminMemosDeleteHandler(w, r)
		return
	}
	if status == enums.MEMO_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoInvalidStatus.Error()))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrTimeParse.Error()))
		return
	}

	if err := s.services.memo.UpdateMemo(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
		idStr,
		title,
		message,
		fileURL,
		status,
		startDate,
		endDate,
		staffIDs,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Memo updated successfully"))
}

func (s *Server) adminMemosDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Delete Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var p forms.AdminMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoDeleteFailed.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoDeleteFailed.Error()))
		return
	}
	if err := s.services.memo.SoftDeleteMemo(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMemoDeleteFailed.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Memo deleted successfully"))
}

func (s *Server) adminMemosSendEmailsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Send Emails Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	var p forms.AdminMemoPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	isSuperuser := false
	if staff, err := s.services.staff.GetCurrentStaff(ctx, currentStaffID); err == nil {
		isSuperuser = staff.UserType == enums.STAFF_USER_TYPE_SUPERUSER.String()
	}

	if err := s.services.memo.SendMemoEmails(ctx, currentStaffID, idStr, isSuperuser); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Notification emails sent successfully"))
}
