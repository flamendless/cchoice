package products

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"

	"go.uber.org/zap"
)

type ProductCategoryServer struct {
	pb.UnimplementedProductCategoryServiceServer
	CtxDB *ctx.Database
}

func productCategoryFromRow(row *cchoice_db.TblProductCategory) *pb.ProductCategory {
	return &pb.ProductCategory{
		Id:          serialize.EncDBID(row.ID),
		ProductId:   serialize.EncDBID(row.ProductID),
		Category:    row.Category.String,
		Subcategory: row.Subcategory.String,
	}
}

func (s *ProductCategoryServer) GetProductCategoryByID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.ProductCategory, error) {
	encid := in.GetId()
	logs.Log().Debug("GetProductCategoryByID", zap.String("encid", encid))

	existingProductCategory, err := s.CtxDB.QueriesRead.GetProductCategoryByID(
		ctx,
		serialize.DecDBID(encid),
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.IDNotFound, err.Error())
	}
	return productCategoryFromRow(&existingProductCategory), nil
}

func (s *ProductCategoryServer) GetProductCategoryByProductID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.ProductCategory, error) {
	encid := in.GetId()
	logs.Log().Debug("GetProductCategoryByProductID", zap.String("encid", encid))

	existingProductCategory, err := s.CtxDB.QueriesRead.GetProductCategoryByProductID(
		ctx,
		serialize.DecDBID(encid),
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.IDNotFound, err.Error())
	}
	return productCategoryFromRow(&existingProductCategory), nil
}
