package handlers

import (
	"cchoice/client/common"
	"cchoice/client/components"
	pb "cchoice/proto"
	"context"
	"net/http"

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
	res, err := h.BrandService.GetBrandLogos(context.Background(), &pb.GetBrandLogosRequest{})
	if err != nil {
		return &common.HandlerRes{Error: err}
	}
	return &common.HandlerRes{
		Component: components.ShopBrands(res.Brands),
	}
}
