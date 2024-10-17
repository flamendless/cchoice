package products

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
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

func (s *ProductCategoryServer) GetProductsByCategoryID(
	ctx context.Context,
	in *pb.GetProductsByCategoryIDRequest,
) (*pb.GetProductsByCategoryIDResponse, error) {
	logs.Log().Debug("GetProductCategoryByID", zap.Int64("category id", in.CategoryId))
	res, err := s.CtxDB.QueriesRead.GetProductsByCategoryID(
		ctx,
		cchoice_db.GetProductsByCategoryIDParams{
			Limit:      in.Limit,
			CategoryID: in.CategoryId,
		},
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
	}

	data := make([]*pb.ProductByCategory, 0, len(res))
	for _, productByCategory := range res {
		data = append(data, &pb.ProductByCategory{
			Id:          serialize.EncDBID(productByCategory.ID),
			CategoryId:  serialize.EncDBID(in.CategoryId),
			Name:        productByCategory.Name,
			Description: productByCategory.Description.String,
			BrandName:   productByCategory.BrandName,
			Thumbnail:   productByCategory.Thumbnail,
			UnitPriceWithVatDisplay: utils.NewMoney(
				productByCategory.UnitPriceWithVat,
				productByCategory.UnitPriceWithVatCurrency,
			).Display(),
		})
	}

	return &pb.GetProductsByCategoryIDResponse{
		Length:   int64(len(res)),
		Products: data,
	}, nil
}
