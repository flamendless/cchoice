package server

import (
	"net/http"

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

func (s *Server) adminQuotationsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Quotations List Page Handler]"
	const page = "/admin/quotations"
	ctx := r.Context()

	if err := compadmin.AdminQuotationsListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminQuotationsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Quotations List Table Handler]"
	const page = "/admin/quotations"
	ctx := r.Context()

	var q forms.AdminQuotationsListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	sortBy, sortDir, err := utils.ParseListingSortQuery(r.URL.Query(), "CREATED_AT", "STATUS")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("sort_by", sortBy), zap.String("sort_dir", sortDir.String()), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	listPage := httputil.PageOrDefault(q.Page, 1)

	serviceQuotations, totalCount, listPage, err := s.services.quotation.GetForListingAdminPaginated(
		ctx,
		q.Search,
		sortBy,
		sortDir,
		listPage,
		constants.DefaultAdminTablePageSize,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	quotations := make([]models.AdminQuotationListItem, 0, len(serviceQuotations))
	for _, q := range serviceQuotations {
		quotations = append(quotations, models.AdminQuotationListItem{
			ID:           s.encoder.Encode(q.ID),
			CustomerName: q.CustomerName,
			Status:       q.Status,
			AssignedTo:   q.AssignedTo,
			TotalItems:   q.TotalItems,
			TotalDisplay: q.TotalDisplay,
			SubmittedAt:  q.SubmittedAt,
			UpdatedAt:    q.UpdatedAt,
		})
	}

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/quotations/table"),
		Include:       "[name='search'],[name='sort_by'],[name='sort_dir']",
		ContentTarget: "#quotations-table-content",
	}

	if err := compadmin.AdminQuotationsTableContent(quotations, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminQuotationsDetailsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Quotations Details Handler]"
	const page = "/admin/quotations"
	ctx := r.Context()

	var p forms.AdminQuotationPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	lines, err := s.services.quotation.GetLinesForAdmin(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	lineItems := make([]models.AdminQuotationLineItem, 0, len(lines))
	for _, line := range lines {
		lineItems = append(lineItems, models.AdminQuotationLineItem{
			BrandName:     line.BrandName,
			ProductSerial: line.ProductSerial,
			Quantity:      line.Quantity,
			TotalPrice:    line.TotalPrice,
			TotalDiscount: line.TotalDiscount,
		})
	}

	if err := compadmin.QuotationDetailsRows(lineItems).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminQuotationsApproveModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Quotations Approve Modal Handler]"
	const page = "/admin/quotations"
	ctx := r.Context()

	var p forms.AdminQuotationPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	track, err := s.services.quotation.GetStatusHistoryForAdmin(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	if track.CurrentStatus != enums.QUOTATION_STATUS_IN_REVIEW {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrQuotationNotApprovable.Error()))
		return
	}

	staffModels, err := s.services.staff.GetAll(ctx, 100)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	modalData := models.AdminQuotationApproveModalData{
		ID:            idStr,
		CurrentStatus: track.CurrentStatus,
		Staff:         staffModels,
	}

	if err := compadmin.QuotationApproveModal(modalData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminQuotationsApproveHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Quotations Approve Handler]"
	const page = "/admin/quotations"
	ctx := r.Context()

	var p forms.AdminQuotationPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	var f forms.AdminQuotationApproveForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	assignedStaffID, err := httputil.RequireEncodedID(s.encoder, f.AssignedStaffID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMissingField.Error()))
		return
	}
	notes := f.Notes
	actingStaffID := s.sessionManager.GetString(ctx, SessionStaffID)

	if err := s.services.quotation.Approve(ctx, actingStaffID, idStr, assignedStaffID, notes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Quotation approved successfully"))
}

func mapQuotationHistoryToModel(history []services.QuotationStatusHistoryEntry) []models.AdminQuotationStatusHistoryEntry {
	result := make([]models.AdminQuotationStatusHistoryEntry, 0, len(history))
	for _, entry := range history {
		result = append(result, models.AdminQuotationStatusHistoryEntry{
			FromStatus: entry.FromStatus,
			ToStatus:   entry.ToStatus,
			StaffName:  entry.StaffName,
			Notes:      entry.Notes,
			CreatedAt:  entry.CreatedAt,
		})
	}
	return result
}
