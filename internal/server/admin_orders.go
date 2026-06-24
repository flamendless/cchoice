package server

import (
	"net/http"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminOrdersListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Orders List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminOrdersListPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/orders", err.Error()))
	}
}

func (s *Server) adminOrdersListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Orders List Table Handler]"
	const page = "/admin/orders"
	ctx := r.Context()

	searchOrderRef := r.URL.Query().Get("search_order_ref")
	sortBy := strings.ToUpper(r.URL.Query().Get("sort_by"))
	sortDir := strings.ToUpper(r.URL.Query().Get("sort_dir"))

	switch sortBy {
	case "", "UPDATED_AT", "CREATED_AT", "STATUS":
	default:
		logs.LogCtx(ctx).Error(logtag, zap.String("sort_by", sortBy), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	switch sortDir {
	case "", "ASC", "DESC":
	default:
		logs.LogCtx(ctx).Error(logtag, zap.String("sort_dir", sortDir), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	serviceOrders, err := s.services.order.GetForListingAdmin(ctx, searchOrderRef, sortBy, sortDir)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	orders := make([]models.AdminOrderListItem, 0, len(serviceOrders))
	for _, o := range serviceOrders {
		orders = append(orders, models.AdminOrderListItem{
			ID:             s.encoder.Encode(o.ID),
			OrderReference: o.OrderReference,
			Status:         o.Status,
			IsPaid:         o.IsPaid,
			CreatedAt:      o.CreatedAt,
			UpdatedAt:      o.UpdatedAt,
		})
	}

	if err := compadmin.AdminOrdersListTable(orders).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminOrdersDetailsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Orders Details Handler]"
	const page = "/admin/orders"
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	details, err := s.services.order.GetDetailsForAdmin(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	lines := make([]models.AdminOrderLineItem, 0, len(details.Lines))
	for _, line := range details.Lines {
		lines = append(lines, models.AdminOrderLineItem{
			Name:        line.Name,
			Serial:      line.Serial,
			Description: line.Description,
			UnitPrice:   line.UnitPrice,
			Quantity:    line.Quantity,
			TotalPrice:  line.TotalPrice,
		})
	}

	if err := compadmin.OrderDetailsRows(models.AdminOrderDetails{
		Customer: models.AdminOrderCustomerInfo{
			Name:    details.Customer.Name,
			Email:   details.Customer.Email,
			Phone:   details.Customer.Phone,
			Address: details.Customer.Address,
		},
		Lines: lines,
	}).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}
