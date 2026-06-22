package server

import (
	"net/http"
	"strings"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminMemosListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminMemosListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/memos", err.Error()))
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

	idStr := chi.URLParam(r, "id")
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
	ctx := r.Context()

	staffList, err := s.services.staff.GetAll(ctx, 100)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/memos", err.Error()))
		return
	}

	currentStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := compadmin.MemoCreateModal(staffList, nil, currentStaffID).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/memos", err.Error()))
	}
}

func (s *Server) adminMemosEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Edit Page Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	memo, err := s.services.memo.GetMemoByID(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to get memo"))
		return
	}
	if memo == nil {
		redirectHX(w, r, utils.URLWithError(page, "Memo not found"))
		return
	}

	recipientIDs, err := s.services.memo.GetRecipientStaffIDs(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	staffList, err := s.services.staff.GetAll(ctx, 0)
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
		redirectHX(w, r, utils.URLWithError(page, "Failed to render edit form"))
	}
}

func (s *Server) adminMemosCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Create Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	title := r.FormValue("title")
	message := r.FormValue("message")
	fileURL := strings.TrimSpace(r.FormValue("file_url"))
	statusStr := r.FormValue("status")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	staffIDs := r.Form["staff_ids"]

	if title == "" || message == "" || statusStr == "" || startDateStr == "" || endDateStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid end date format"))
		return
	}

	if err := services.ValidateMemoDates(startDate, endDate); err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	status := enums.ParseMemoStatusToEnum(statusStr)
	if status == enums.MEMO_STATUS_UNDEFINED || status == enums.MEMO_STATUS_DELETED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid status"))
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

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to parse form"))
		return
	}

	idStr := chi.URLParam(r, "id")
	title := r.FormValue("title")
	message := r.FormValue("message")
	fileURL := strings.TrimSpace(r.FormValue("file_url"))
	statusStr := r.FormValue("status")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	staffIDs := r.Form["staff_ids"]

	if idStr == "" || title == "" || message == "" || statusStr == "" || startDateStr == "" || endDateStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "All fields are required"))
		return
	}

	status := enums.ParseMemoStatusToEnum(statusStr)
	if status == enums.MEMO_STATUS_DELETED {
		s.adminMemosDeleteHandler(w, r)
		return
	}
	if status == enums.MEMO_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid status"))
		return
	}

	startDate, err := time.Parse(constants.DateLayoutISO, startDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(constants.DateLayoutISO, endDateStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid end date format"))
		return
	}

	if err := services.ValidateMemoDates(startDate, endDate); err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
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

	idStr := chi.URLParam(r, "id")
	if err := s.services.memo.SoftDeleteMemo(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Failed to delete memo"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Memo deleted successfully"))
}

func (s *Server) adminMemosSendEmailsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Memos Send Emails Handler]"
	const page = "/admin/memos"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
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
