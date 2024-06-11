package product

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/domains/grpc"
	"cchoice/internal/logs"
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
		ID:          row.ID,
		ProductId:   row.ProductID,
		Category:    row.Category.String,
		Subcategory: row.Subcategory.String,
	}
}

func (s *ProductCategoryServer) GetProductCategoryByID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.ProductCategory, error) {
	id := in.GetId()
	logs.Log().Debug("GetProductCategoryByID", zap.Int64("id", id))

	existingProductCategory, err := s.CtxDB.QueriesRead.GetProductCategoryByID(ctx, id)
	if err != nil {
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
	}
	return productCategoryFromRow(&existingProductCategory), nil
}

func (s *ProductCategoryServer) GetProductCategoryByProductID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.ProductCategory, error) {
	id := in.GetId()
	logs.Log().Debug("GetProductCategoryByProductID", zap.Int64("id", id))

	existingProductCategory, err := s.CtxDB.QueriesRead.GetProductCategoryByProductID(ctx, id)
	if err != nil {
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
	}
	return productCategoryFromRow(&existingProductCategory), nil
}
