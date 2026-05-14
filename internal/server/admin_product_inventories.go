package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminProductInventoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Page Handler]"
	ctx := r.Context()

	brandsRes, err := requests.GetBrandsForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminBrandsCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		brandsRes = []queries.GetBrandsForProductCreateRow{}
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   s.encoder.Encode(b.ID),
			Name: b.Name,
		})
	}

	if err := compadmin.AdminProductInventoriesPage(brands).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/product-inventories", err.Error()))
		return
	}
}

func (s *Server) adminProductInventoriesTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Table Handler]"
	const page = "/admin"
	ctx := r.Context()

	searchSerial := r.URL.Query().Get("search_serial")
	searchBrand := r.URL.Query().Get("search_brand")

	statusStr := r.URL.Query().Get("product_status")
	productStatus := enums.ParseProductStatusToEnum(statusStr)
	if statusStr != "" && productStatus == enums.PRODUCT_STATUS_UNDEFINED {
		logs.LogCtx(ctx).Error(logtag, zap.String("status", statusStr), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	stocksInStr := r.URL.Query().Get("stocks_in")
	stocksIn := enums.ParseStocksInToEnum(stocksInStr)
	if stocksInStr != "" && stocksIn == enums.STOCKS_IN_UNDEFINED {
		logs.LogCtx(ctx).Error(logtag, zap.String("stocks_in", stocksInStr), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	inventories, err := s.services.productInventory.GetListingForAdmin(
		ctx,
		searchSerial,
		searchBrand,
		productStatus,
		stocksIn,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}

	if err := compadmin.AdminProductInventoriesTable(inventories).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}
