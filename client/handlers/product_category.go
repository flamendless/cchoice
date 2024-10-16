package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/errs"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"net/http"
	"strconv"

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
	qlimit := r.URL.Query().Get("limit")
	if qlimit == "" {
		qlimit = "100"
	}
	limit, err := strconv.Atoi(qlimit)
	if err != nil {
		return &common.HandlerRes{Error: errs.ERR_INVALID_PARAMS}
	}

	res, err := h.ProductCategoryService.GetProductCategoriesByPromoted(
		context.TODO(),
		&pb.GetProductCategoriesByPromotedRequest{
			Limit:              int64(limit),
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
