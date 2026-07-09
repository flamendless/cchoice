package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) adminProductInventoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Page Handler]"
	const page = "/admin/product-inventories"
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
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) adminProductInventoriesTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventories Table Handler]"
	const page = "/admin/product-inventories"
	ctx := r.Context()

	var q forms.AdminProductInventoriesListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	searchSerial := q.SearchSerial
	searchBrand := q.SearchBrand

	statusStr := q.ProductStatus
	productStatus := enums.ParseProductStatusToEnum(statusStr)
	if statusStr != "" && productStatus == enums.PRODUCT_STATUS_UNDEFINED {
		logs.LogCtx(ctx).Error(logtag, zap.String("status", statusStr), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	stocksInStr := q.StocksIn
	stocksIn := enums.ParseStocksInToEnum(stocksInStr)
	if stocksInStr != "" && stocksIn == enums.STOCKS_IN_UNDEFINED {
		logs.LogCtx(ctx).Error(logtag, zap.String("stocks_in", stocksInStr), zap.Error(errs.ErrEnumInvalid))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	listPage := httputil.PageOrDefault(q.Page, 1)

	inventories, totalCount, listPage, err := s.services.productInventory.GetListingForAdminPaginated(
		ctx,
		searchSerial,
		searchBrand,
		productStatus,
		stocksIn,
		listPage,
		constants.DefaultAdminTablePageSize,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/product-inventories/table"),
		Include:       "[name='search_serial'],[name='search_brand'],[name='product_status'],[name='stocks_in']",
		ContentTarget: "#inventories-table-content",
	}

	if err := compadmin.AdminProductInventoriesTableContent(inventories, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminProductInventoryUpdateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventory Update Modal Handler]"
	const page = "/admin/product-inventories"
	ctx := r.Context()

	var p forms.AdminProductInventoryPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrDecode.Error()))
		return
	}
	inventoryID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrDecode.Error()))
		return
	}
	decoded := s.encoder.Decode(inventoryID)
	if decoded == encode.INVALID {
		logs.LogCtx(ctx).Error(logtag, zap.String("inventory_id", inventoryID), zap.Error(errs.ErrDecode))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrDecode.Error()))
		return
	}

	inv, err := s.dbRO.GetQueries().GetProductInventoryByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logs.LogCtx(ctx).Error(logtag, zap.Int64("decoded", decoded), zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInventoryNotFound.Error()))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	data := models.AdminProductInventoryListItem{
		ID:        inventoryID,
		ProductID: s.encoder.Encode(inv.ProductID),
		StocksIn:  enums.ParseStocksInToEnum(inv.StocksIn),
		Stocks:    inv.Stocks,
	}

	if err := compadmin.InventoryUpdateModal(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminProductInventoryUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Product Inventory Update Handler]"
	const page = "/admin/product-inventories"
	ctx := r.Context()

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if staffID == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrForbidden.Error()))
		return
	}

	var p forms.AdminProductInventoryPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	inventoryID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	var f forms.AdminProductInventoryUpdateForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	qtyStr := f.Qty
	stocksInStr := f.StocksIn

	stocksIn := enums.ParseStocksInToEnum(stocksInStr)
	if stocksIn == enums.STOCKS_IN_UNDEFINED {
		logs.LogCtx(ctx).Error(logtag, zap.String("stocks_in", stocksInStr), zap.Error(errs.ErrInvalidInput))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrProductInvalidStocksIn.Error()))
		return
	}

	qty, err := strconv.ParseInt(qtyStr, 10, 64)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("qty", qtyStr), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrParseInt.Error()))
		return
	}

	if err := s.services.productInventory.UpdateByID(ctx, staffID, inventoryID, qty, stocksIn); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Inventory updated successfully"))
}
