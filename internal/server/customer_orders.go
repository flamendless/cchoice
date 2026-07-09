package server

import (
	"net/http"

	compcustomer "cchoice/cmd/web/components/customers"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) customerOrdersListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Orders List Page Handler]"
	ctx := r.Context()

	if err := compcustomer.CustomerOrdersListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) customerOrdersListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Orders List Table Handler]"
	const page = "/customer/orders"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	var q forms.CustomerOrdersListQuery
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

	serviceOrders, totalCount, listPage, err := s.services.order.GetForListingCustomerPaginated(
		ctx,
		customerIDStr,
		q.SearchOrderRef,
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

	orders := make([]models.CustomerOrderListItem, 0, len(serviceOrders))
	for _, o := range serviceOrders {
		orders = append(orders, models.CustomerOrderListItem{
			ID:             s.encoder.Encode(o.ID),
			OrderReference: o.OrderReference,
			Status:         o.Status,
			IsPaid:         o.IsPaid,
			OrderedAt:      o.OrderedAt,
			EarnedCPoints:  utils.FormatEarnedCPoints(o.EarnedCPoints),
		})
	}

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/customer/orders/table"),
		Include:       "[name='search_order_ref'],[name='sort_by'],[name='sort_dir']",
		ContentTarget: "#customer-orders-table-content",
	}

	if err := compcustomer.CustomerOrdersTableContent(orders, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) customerOrderDetailPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Customer Order Detail Page Handler]"
	const page = "/customer/orders"
	ctx := r.Context()

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	var p forms.CustomerOrderPath
	if err := httputil.BindPath(r, &p); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)), http.StatusSeeOther)
		return
	}
	orderIDStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URLWithError(page, err.Error()), http.StatusSeeOther)
		return
	}

	profile, err := s.services.customer.BuildProfile(ctx, customerIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URLWithError(page, err.Error()), http.StatusSeeOther)
		return
	}

	details, err := s.services.order.GetDetailsForCustomer(ctx, customerIDStr, orderIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URLWithError(page, err.Error()), http.StatusSeeOther)
		return
	}

	trackData, err := s.services.order.GetStatusHistoryForCustomer(ctx, customerIDStr, orderIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URLWithError(page, err.Error()), http.StatusSeeOther)
		return
	}

	pageData := models.CustomerOrderDetailPageData{
		Profile: profile,
		Details: mapAdminOrderDetails(details),
		Track:   mapAdminOrderTrackData(orderIDStr, trackData),
	}

	if err := compcustomer.CustomerOrderDetailPage(pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mapAdminOrderTrackData(orderIDStr string, trackData *services.OrderAdminTrackData) models.AdminOrderTrackModalData {
	history := make([]models.AdminOrderStatusHistoryEntry, 0, len(trackData.History))
	for _, entry := range trackData.History {
		history = append(history, models.AdminOrderStatusHistoryEntry{
			FromStatus: entry.FromStatus,
			ToStatus:   entry.ToStatus,
			StaffName:  entry.StaffName,
			Notes:      entry.Notes,
			CreatedAt:  entry.CreatedAt,
		})
	}

	return models.AdminOrderTrackModalData{
		ID:             orderIDStr,
		OrderReference: trackData.OrderReference,
		CurrentStatus:  trackData.CurrentStatus,
		History:        history,
		FlowSteps:      trackData.FlowSteps,
	}
}
