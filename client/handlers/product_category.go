package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"net/http"

	"go.uber.org/zap"
)

type ProductCategoryService interface {
	pb.ProductCategoryServiceClient
}

type ProductCategoryHandler struct {
	Logger                 *zap.Logger
	ProductCategoryService ProductCategoryService
}

func NewProductCategoryHandler(
	logger *zap.Logger,
	productCategoryService ProductCategoryService,
) ProductCategoryHandler {
	return ProductCategoryHandler{
		Logger:                 logger,
		ProductCategoryService: productCategoryService,
	}
}

func (h ProductCategoryHandler) ProductsCategories(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	limit, err := utils.GetLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return &common.HandlerRes{Error: err}
	}

	res, err := h.ProductCategoryService.GetProductCategoriesByPromoted(
		context.TODO(),
		&pb.GetProductCategoriesByPromotedRequest{
			Limit:              limit,
			PromotedAtHomepage: true,
		},
	)
	if err != nil {
		return &common.HandlerRes{Error: err, StatusCode: http.StatusBadRequest}
	}

	data := make([]*common.ShopProductCategory, 0, res.Length)
	for _, productCategory := range res.ProductsCategories {
		data = append(data, &common.ShopProductCategory{
			ID:            productCategory.Id,
			Category:      utils.SlugToTitle(productCategory.Category),
			ProductsCount: productCategory.ProductsCount,
		})
	}

	return &common.HandlerRes{
		Component: components.ShopProductsCategories(data),
	}
}

func (h ProductCategoryHandler) ProductCategoryProducts(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	id := r.PathValue("id")
	if id == "" {
		return &common.HandlerRes{Error: errs.ERR_INVALID_RESOURCE}
	}

	limit, err := utils.GetLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return &common.HandlerRes{Error: err}
	}

	ctx := context.TODO()
	res, err := h.ProductCategoryService.GetProductsByCategoryID(
		ctx,
		&pb.GetProductsByCategoryIDRequest{
			Limit:      limit,
			CategoryId: serialize.DecDBID(id),
		},
	)
	if err != nil {
		return &common.HandlerRes{Error: errs.ERR_INVALID_RESOURCE}
	}

	return &common.HandlerRes{
		Component: components.ShopCategoryProducts(id, res.Products),
	}
}
