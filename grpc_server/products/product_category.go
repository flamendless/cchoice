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
) (*pb.GetProductCategoriesByPromotedResponse, error) {
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

	productsCategories := make([]*pb.ProductsCategories, 0, len(promotedProductCategories))
	for _, pc := range promotedProductCategories {
		productsCategories = append(productsCategories, &pb.ProductsCategories{
			Id:            serialize.EncDBID(pc.ID),
			Category:      pc.Category.String,
			ProductsCount: pc.ProductsCount,
		})
	}

	return &pb.GetProductCategoriesByPromotedResponse{
		Length:             int64(len(promotedProductCategories)),
		ProductsCategories: productsCategories,
	}, nil
}
