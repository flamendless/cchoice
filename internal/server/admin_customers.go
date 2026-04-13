package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCustomersListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Customers List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminCustomersListPage("Customers").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/customers", err.Error()))
		return
	}
}

func (s *Server) adminCustomersListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Customers List Table Handler]"
	const page = "/admin/superuser/customers"
	ctx := r.Context()

	email := r.URL.Query().Get("email")
	customerType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	customers, err := s.services.customer.FilterCustomers(ctx, email, customerType, status)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	customerList := make([]models.AdminCustomerListItem, 0, len(customers))
	for _, c := range customers {
		customerList = append(customerList, models.AdminCustomerListItem(c))
	}

	if err := compadmin.AdminCustomersListTable(customerList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}
