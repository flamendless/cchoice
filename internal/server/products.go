package server

import (
	compproduct "cchoice/cmd/web/components/product"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddProductHandlers(s *Server, r chi.Router) {
	r.Get("/product/{slug}", s.productPageHandler)
}

func (s *Server) productPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Product Page Handler]"
	const page = "/"
	ctx := r.Context()

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}

	productData, err := s.services.product.GetProductPage(ctx, slug)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("slug", slug))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	decodedProductID := s.encoder.Decode(productData.ProductID)
	if decodedProductID == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}

	// TODO: Query related products by category

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
