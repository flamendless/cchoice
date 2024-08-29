package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

type BrandService interface {
	pb.BrandServiceClient
}

type BrandHandler struct {
	Logger       *zap.Logger
	BrandService BrandService
}

func NewBrandHandler(
	logger *zap.Logger,
	brandService BrandService,
) BrandHandler {
	return BrandHandler{
		Logger:       logger,
		BrandService: brandService,
	}
}

func (h BrandHandler) BrandLogos(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	qlimit := r.URL.Query().Get("limit")
	if qlimit == "" {
		qlimit = "100"
	}

	limit, err := strconv.Atoi(qlimit)
	if err != nil {
		return &common.HandlerRes{Error: errs.ERR_INVALID_PARAMS}
	}

	res, err := h.BrandService.GetBrandLogos(context.Background(), &pb.GetBrandLogosRequest{
		Limit: int64(limit),
	})
	if err != nil {
		return &common.HandlerRes{Error: err}
	}
	return &common.HandlerRes{
		Component: components.ShopBrands(res.Brands),
	}
}

func (h BrandHandler) BrandPage(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	id := r.PathValue("id")
	if id == "" {
		return &common.HandlerRes{Error: errs.ERR_INVALID_RESOURCE}
	}

	res, err := h.BrandService.GetBrand(context.Background(), &pb.GetBrandRequest{
		Id: id,
	})
	if err != nil || res.Brand == nil {
		return &common.HandlerRes{Error: err}
	}

	return &common.HandlerRes{
		// Component: components.ShopBrands(res.Brand),
	}
}
