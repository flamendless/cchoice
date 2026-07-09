package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCustomersListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Customers List Page Handler]"
	const page = "/admin/superuser/customers"
	ctx := r.Context()

	if err := compadmin.AdminCustomersListPage("Customers").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminCustomersListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Customers List Table Handler]"
	const page = "/admin/superuser/customers"
	ctx := r.Context()

	var q forms.AdminCustomersFilterQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	customers, err := s.services.customer.FilterCustomers(ctx, q.Email, q.Type, q.Status)
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
