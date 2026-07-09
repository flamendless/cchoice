package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"cchoice/cmd/web/components"
	compproduct "cchoice/cmd/web/components/product"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"
)

func AddProductHandlers(s *Server, r chi.Router) {
	r.Get("/product/{slug}", s.productPageHandler)
}

func (s *Server) productPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Product Page Handler]"
	const page = "/"
	ctx := r.Context()

	var pathReq forms.ProductSlugPath
	if err := httputil.BindPath(r, &pathReq); err != nil {
		s.renderProductNotFound(w, r, page)
		return
	}
	slug := pathReq.Slug

	productData, err := s.services.product.GetForPage(ctx, slug)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("slug", slug))
		s.renderProductNotFound(w, r, page)
		return
	}

	decodedProductID := s.encoder.Decode(productData.ProductID)
	if decodedProductID == encode.INVALID {
		s.renderProductNotFound(w, r, page)
		return
	}

	if err := compproduct.ProductPage(*productData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
}

func (s *Server) renderProductNotFound(w http.ResponseWriter, r *http.Request, fallbackURL string) {
	if isHTMX(r) {
		redirectHX(w, r, utils.URLWithError(fallbackURL, errs.ErrNotFound.Error()))
		return
	}

	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	if err := components.NotPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error("[Product Page Handler]", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
