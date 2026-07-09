package server

import (
	"net/http"

	compcustomer "cchoice/cmd/web/components/customers"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) customerQuotationsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Quotations List Page Handler]"
	ctx := r.Context()

	if err := compcustomer.CustomerQuotationsListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerQuotationsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Quotations List Table Handler]"
	const page = "/customer/quotations"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	var q forms.CustomerQuotationsListQuery
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

	serviceQuotations, totalCount, listPage, err := s.services.quotation.GetForListingCustomerPaginated(
		ctx,
		customerIDStr,
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

	quotations := make([]models.CustomerQuotationListItem, 0, len(serviceQuotations))
	for _, q := range serviceQuotations {
		quotations = append(quotations, models.CustomerQuotationListItem{
			ID:           s.encoder.Encode(q.ID),
			Status:       q.Status,
			TotalItems:   q.TotalItems,
			TotalDisplay: q.TotalDisplay,
			SubmittedAt:  q.SubmittedAt,
		})
	}

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/customer/quotations/table"),
		Include:       "[name='sort_by'],[name='sort_dir']",
		ContentTarget: "#customer-quotations-table-content",
	}

	if err := compcustomer.CustomerQuotationsTableContent(quotations, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) customerQuotationDetailPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Quotation Detail Page Handler]"
	const page = "/customer/quotations"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	var p forms.CustomerQuotationDetailPath
	if err := httputil.BindPath(r, &p); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	quotationIDStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	detail, err := s.services.quotation.GetDetailForCustomer(ctx, customerIDStr, quotationIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	lines := make([]models.AdminQuotationLineItem, 0, len(detail.Lines))
	for _, line := range detail.Lines {
		lines = append(lines, models.AdminQuotationLineItem{
			BrandName:     line.BrandName,
			ProductSerial: line.ProductSerial,
			Quantity:      line.Quantity,
			TotalPrice:    line.TotalPrice,
			TotalDiscount: line.TotalDiscount,
		})
	}

	pageData := models.CustomerQuotationDetailPageData{
		ID:             s.encoder.Encode(detail.ID),
		Status:         detail.Status,
		SubmittedAt:    detail.SubmittedAt,
		UpdatedAt:      detail.UpdatedAt,
		Lines:          lines,
		TotalItems:     detail.TotalItems,
		TotalPrice:     detail.TotalPrice,
		TotalDiscounts: detail.TotalDiscounts,
		Total:          detail.Total,
		History:        mapQuotationHistoryToModel(detail.Track.History),
		FlowSteps:      detail.Track.FlowSteps,
		CurrentStatus:  detail.Track.CurrentStatus,
	}

	if err := compcustomer.CustomerQuotationDetailPage(pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
