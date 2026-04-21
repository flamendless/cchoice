package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminProductInventoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminProductInventoriesPage("Product Inventories").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/product-inventories", err.Error()))
		return
	}
}

func (s *Server) adminProductInventoriesTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Table Handler]"
	const page = "/admin"
	ctx := r.Context()

	inventories, err := s.services.productInventory.GetListingForAdmin(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}

	if err := compadmin.AdminProductInventoriesTable(inventories).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}
