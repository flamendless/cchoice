package server

import (
	"fmt"
	"net/http"
	"strings"

	compshop "cchoice/cmd/web/components/shop"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddProductCategoriesHandlers(s *Server, r chi.Router) {
	r.Get("/product-categories/side-panel/list", s.categoriesSidePanelHandler)
	r.Get("/product-categories/sections", s.categorySectionHandler)
	r.Get("/product-categories/products/batch", s.categoryProductsBatchHandler)
	r.Get("/product-categories/{category_id}/products", s.categoryProductsHandler)
}

func (s *Server) categoriesSidePanelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Categories Side Panel Handler]"
	ctx := r.Context()

	categories, err := requests.GetCategoriesSidePanel(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_categories_side_panel"),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := compshop.CategoriesSidePanelList(categories).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) categorySectionHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Categories Section Handler]"
	ctx := r.Context()

	var req forms.CategorySectionQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}

	page := req.Page
	limit := req.EffectiveLimit()

	res, err := requests.GetCategorySectionHandler(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		s.encoder,
		fmt.Appendf([]byte{}, "categorySectionHandler_p%d_l%d", page, limit),
		page,
		limit,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := compshop.CategorySection(page, res, nil, true).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ParseCategoryProductBatchIDs(idsParam string) ([]string, error) {
	if strings.TrimSpace(idsParam) == "" {
		return nil, errs.ErrInvalidParams
	}

	parts := strings.Split(idsParam, ",")
	ids := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
		if len(ids) > constants.DefaultShopBatchProductSections {
			return nil, errs.ErrInvalidParams
		}
	}

	if len(ids) == 0 {
		return nil, errs.ErrInvalidParams
	}
	return ids, nil
}

func (s *Server) categoryProductsBatchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Category Products Batch Handler]"
	ctx := r.Context()

	var req forms.CategoryProductsBatchQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}

	ids, err := ParseCategoryProductBatchIDs(req.IDs)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	for _, id := range ids {
		if _, err := httputil.RequireEncodedID(s.encoder, id); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}
	}

	var brandID int64
	filters := GetHomePageFilters(ctx, s.sessionManager)
	if filters.BrandID != "" {
		brandID = s.encoder.Decode(filters.BrandID)
	}

	sections := make([]models.CategorySectionProducts, 0, len(ids))
	for _, id := range ids {
		sectionProducts, err := s.loadCategorySectionProducts(ctx, id, brandID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("category_id", id))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sections = append(sections, sectionProducts)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := compshop.CategorySectionBatchResponse(sections).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) categoryProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Category Products Handler]"
	ctx := r.Context()

	var pathReq forms.CategoryProductsPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}
	if _, err := httputil.RequireEncodedID(s.encoder, pathReq.CategoryID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	var brandID int64
	filters := GetHomePageFilters(ctx, s.sessionManager)
	if filters.BrandID != "" {
		brandID = s.encoder.Decode(filters.BrandID)
	}

	categorySectionProducts, err := s.loadCategorySectionProducts(ctx, pathReq.CategoryID, brandID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := compshop.CategorySectionProductsInner(categorySectionProducts.WithHighPriority(true)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
