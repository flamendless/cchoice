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
	r.Get("/product/{id}", s.productPageHandler)
}

func (s *Server) productPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Product Page Handler]"
	const page = "/"
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	if productID == "" {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}

	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrNotFound.Error()))
		return
	}

	productData, err := s.services.product.GetProductPage(ctx, productID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("product_id", productID))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	categoryID, err := s.services.product.GetProductCategoryID(ctx, decodedProductID)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("product_id", productID))
	} else {
		relatedProducts, err := s.services.product.GetRelatedProducts(ctx, categoryID, decodedProductID)
		if err != nil {
			logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("product_id", productID))
		} else {
			productData.RelatedProducts = relatedProducts
		}
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
