package server

import (
	"errors"
	"net/http"

	"cchoice/cmd/web/components"
	compshop "cchoice/cmd/web/components/shop"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddCategoryPageHandlers(s *Server, r chi.Router) {
	r.Get("/categories/{category}", s.categoryPageHandler)
	r.Get("/categories/{category}/{subcategory}", s.categorySubcategoryPageHandler)
}

func (s *Server) categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Category Page Handler]"

	var pathReq forms.CategoryPagePath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		s.renderCategoryNotFound(w, r)
		return
	}

	s.renderCategoryPage(w, r, logtag, pathReq.Category, "")
}

func (s *Server) categorySubcategoryPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Category Subcategory Page Handler]"

	var pathReq forms.CategorySubcategoryPagePath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		s.renderCategoryNotFound(w, r)
		return
	}

	s.renderCategoryPage(w, r, logtag, pathReq.Category, pathReq.Subcategory)
}

func (s *Server) renderCategoryPage(
	w http.ResponseWriter,
	r *http.Request,
	logtag string,
	categorySlug string,
	subcategorySlug string,
) {
	ctx := r.Context()

	pageData, err := s.services.productCategory.GetCategoryPageData(
		ctx,
		categorySlug,
		subcategorySlug,
		s.GetCDNURL,
	)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			s.renderCategoryNotFound(w, r)
			return
		}
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("category", categorySlug),
			zap.String("subcategory", subcategorySlug),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData.ThemeCSS = s.activeThemeCSS(ctx, logtag)

	if err := compshop.CategoryPage(*pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderCategoryNotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	if err := components.NotPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error("[Category Page Handler]", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
