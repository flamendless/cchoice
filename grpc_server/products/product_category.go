package products

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"
	"database/sql"

	"go.uber.org/zap"
)

type ProductCategoryServer struct {
	pb.UnimplementedProductCategoryServiceServer
	CtxDB *ctx.Database
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
	return &pb.ProductCategory{
		Id:          serialize.EncDBID(existingProductCategory.ID),
		Category:    existingProductCategory.Category.String,
		Subcategory: existingProductCategory.Subcategory.String,
	}, nil
}

func (s *ProductCategoryServer) GetProductCategoriesByPromoted(
	ctx context.Context,
	in *pb.GetProductCategoriesByPromotedRequest,
) (*pb.ProductCategories, error) {
	logs.Log().Debug("GetProductCategoriesByPromoted", zap.Bool("promoted at homepage", in.PromotedAtHomepage))
	promotedProductCategories, err := s.CtxDB.QueriesRead.GetProductCategoriesByPromoted(
		ctx,
		cchoice_db.GetProductCategoriesByPromotedParams{
			Limit: in.Limit,
			PromotedAtHomepage: sql.NullBool{
				Bool:  in.PromotedAtHomepage,
				Valid: true,
			},
		},
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
	}

	productCategories := make([]*pb.ProductCategory, 0, len(promotedProductCategories))
	for _, pc := range promotedProductCategories {
		productCategories = append(productCategories, &pb.ProductCategory{
			Id:          serialize.EncDBID(pc.ID),
			Category:    pc.Category.String,
			Subcategory: pc.Subcategory.String,
		})
	}

	return &pb.ProductCategories{
		Length:          int64(len(promotedProductCategories)),
		ProductCategory: productCategories,
	}, nil
}
