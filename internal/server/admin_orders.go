package server

import (
	"net/http"
	"strconv"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/services"
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

	searchOrderRef := strings.TrimSpace(r.URL.Query().Get("search_order_ref"))
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

	listPage := 1
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil && parsed > 0 {
			listPage = parsed
		}
	}

	serviceOrders, totalCount, listPage, err := s.services.order.GetForListingAdminPaginated(
		ctx,
		searchOrderRef,
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

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/orders/table"),
		Include:       "[name='search_order_ref'],[name='sort_by'],[name='sort_dir']",
		ContentTarget: "#orders-table-content",
	}

	if err := compadmin.AdminOrdersTableContent(orders, pagination).Render(ctx, w); err != nil {
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

	if err := compadmin.OrderDetailsRows(mapAdminOrderDetails(details)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func mapAdminOrderDetails(details *services.OrderAdminDetails) models.AdminOrderDetails {
	lines := make([]models.AdminOrderLineItem, 0, len(details.Lines))
	for _, line := range details.Lines {
		lines = append(lines, models.AdminOrderLineItem{
			ThumbnailURL: line.ThumbnailURL,
			Name:         line.Name,
			Serial:       line.Serial,
			UnitPrice:    line.UnitPrice,
			Quantity:     line.Quantity,
			TotalPrice:   line.TotalPrice,
		})
	}

	return models.AdminOrderDetails{
		Order: models.AdminOrderInfo{
			OrderReference: details.Order.OrderReference,
			Status:         details.Order.Status,
			Notes:          details.Order.Notes,
			Remarks:        details.Order.Remarks,
			CreatedAt:      details.Order.CreatedAt,
			UpdatedAt:      details.Order.UpdatedAt,
		},
		Payment: models.AdminOrderPaymentInfo{
			Gateway:         details.Payment.Gateway,
			Status:          details.Payment.Status,
			ReferenceNumber: details.Payment.ReferenceNumber,
			PaymentMethod:   details.Payment.PaymentMethod,
			TotalAmount:     details.Payment.TotalAmount,
			PaidAt:          details.Payment.PaidAt,
			Description:     details.Payment.Description,
			MetadataNotes:   details.Payment.MetadataNotes,
			MetadataRemarks: details.Payment.MetadataRemarks,
			CustomerNumber:  details.Payment.CustomerNumber,
		},
		Shipping: models.AdminOrderShippingInfo{
			AdminOrderAddressInfo: models.AdminOrderAddressInfo{
				Line1:            details.Shipping.Line1,
				Line2:            details.Shipping.Line2,
				City:             details.Shipping.City,
				State:            details.Shipping.State,
				PostalCode:       details.Shipping.PostalCode,
				Country:          details.Shipping.Country,
				FormattedAddress: details.Shipping.FormattedAddress,
			},
			Service:        details.Shipping.Service,
			OrderID:        details.Shipping.OrderID,
			TrackingNumber: details.Shipping.TrackingNumber,
			ETA:            details.Shipping.ETA,
		},
		Billing: models.AdminOrderAddressInfo{
			Line1:            details.Billing.Line1,
			Line2:            details.Billing.Line2,
			City:             details.Billing.City,
			State:            details.Billing.State,
			PostalCode:       details.Billing.PostalCode,
			Country:          details.Billing.Country,
			FormattedAddress: details.Billing.FormattedAddress,
		},
		Customer: models.AdminOrderCustomerInfo{
			Name:  details.Customer.Name,
			Email: details.Customer.Email,
			Phone: details.Customer.Phone,
		},
		Summary: models.AdminOrderAmountSummary{
			Subtotal: details.Summary.Subtotal,
			Shipping: details.Summary.Shipping,
			Discount: details.Summary.Discount,
			Total:    details.Summary.Total,
		},
		Lines: lines,
	}
}
