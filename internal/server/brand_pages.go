package server

import (
	"errors"
	"net/http"

	"cchoice/cmd/web/components"
	compshop "cchoice/cmd/web/components/shop"
	"cchoice/cmd/web/models"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddBrandPageHandlers(s *Server, r chi.Router) {
	r.Get("/brands", s.brandsListingPageHandler)
	r.Get("/brands/{brand}/sections/{section}", s.brandPrioritySectionHandler)
	r.Get("/brands/{brand}/categories/{category_id}/products", s.brandCategoryProductsHandler)
	r.Get("/brands/{brand}", s.brandPageHandler)
}

func (s *Server) brandsListingPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brands Listing Page Handler]"
	ctx := r.Context()

	pageData, err := s.services.brandPage.GetBrandsListingPageData(ctx, s.GetBrandLogoCDNURL)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData.ThemeCSS = s.activeThemeCSS(ctx, logtag)

	if err := compshop.BrandsListingPage(*pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) brandPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brand Page Handler]"

	var pathReq forms.BrandPagePath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		s.renderBrandNotFound(w, r)
		return
	}

	ctx := r.Context()

	pageData, err := s.services.brandPage.GetBrandPageData(
		ctx,
		pathReq.Brand,
		s.GetCDNURL,
		s.GetBrandLogoCDNURL,
	)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			s.renderBrandNotFound(w, r)
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.String("brand", pathReq.Brand), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData.ThemeCSS = s.activeThemeCSS(ctx, logtag)

	if err := compshop.BrandPage(*pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) brandPrioritySectionHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brand Priority Section Handler]"
	ctx := r.Context()

	var pathReq forms.BrandPrioritySectionPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}

	products, err := s.services.brandPage.GetBrandPrioritySectionProducts(
		ctx,
		pathReq.Brand,
		pathReq.Section,
		s.GetCDNURL,
	)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sectionData := modelsCategorySectionProducts(pathReq.Section, products)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := compshop.CategorySectionProductsInner(sectionData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) brandCategoryProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brand Category Products Handler]"
	ctx := r.Context()

	var pathReq forms.BrandCategoryProductsPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}

	sectionProducts, err := s.services.brandPage.GetBrandCategoryProducts(
		ctx,
		pathReq.Brand,
		pathReq.CategoryID,
		s.GetCDNURL,
	)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := compshop.CategorySectionProductsInner(sectionProducts).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderBrandNotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	if err := components.NotPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error("[Brand Page Handler]", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func modelsCategorySectionProducts(section string, products []models.CategorySectionProduct) models.CategorySectionProducts {
	title := section
	switch section {
	case "best-selling":
		title = "Best Selling"
	case "highest-discount":
		title = "Highest Discount"
	}

	return models.CategorySectionProducts{
		ID:          section,
		Category:    title,
		Subcategory: title,
		Products:    products,
	}
}
