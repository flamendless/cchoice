package products

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	CtxDB *ctx.Database
}

type ProductsRow cchoice_db.GetProductsRow

func (row ProductsRow) ToPBProduct() *pb.Product {
	moneyWithoutVat := utils.NewMoney(row.UnitPriceWithoutVat, row.UnitPriceWithoutVatCurrency)
	moneyWithVat := utils.NewMoney(row.UnitPriceWithVat, row.UnitPriceWithVatCurrency)

	return &pb.Product{
		Id:          serialize.EncDBID(row.ID),
		Serial:      row.Serial,
		Name:        row.Name,
		Description: row.Description.String,
		Brand: &pb.Brand{
			Id:   serialize.EncDBID(row.BrandID),
			Name: row.BrandName,
		},
		Status: enums.StringToPBEnum(
			row.Status,
			pb.ProductStatus_ProductStatus_value,
			pb.ProductStatus_UNDEFINED,
		),
		ProductCategory: &pb.ProductCategory{
			Category:    row.Category.String,
			Subcategory: row.Subcategory.String,
		},
		ProductSpecs: &pb.ProductSpecs{
			Colours:       row.Colours.String,
			Sizes:         row.Sizes.String,
			Segmentation:  row.Segmentation.String,
			PartNumber:    row.PartNumber.String,
			Power:         row.Power.String,
			Capacity:      row.Capacity.String,
			ScopeOfSupply: row.ScopeOfSupply.String,
		},
		UnitPriceWithoutVatDisplay: moneyWithoutVat.Display(),
		UnitPriceWithVatDisplay:    moneyWithVat.Display(),
		Metadata: &pb.Metadata{
			CreatedAt: timestamppb.New(row.CreatedAt),
			UpdatedAt: timestamppb.New(row.UpdatedAt),
			DeletedAt: timestamppb.New(row.DeletedAt),
		},
	}
}

func (s *ProductServer) GetProductByID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.Product, error) {
	encid := in.GetId()
	existingProduct, err := s.CtxDB.QueriesRead.GetProductByID(
		ctx,
		serialize.DecDBID(encid),
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.IDNotFound, err.Error())
	}

	res := ProductsRow(existingProduct).ToPBProduct()
	return res, nil
}

func (s *ProductServer) ListProductsByProductStatus(
	ctx context.Context,
	in *pb.ProductStatusRequest,
) (*pb.ProductsResponse, error) {
	status := in.GetStatus()
	sortBy := in.GetSortBy()

	products := make([]*pb.Product, 0, 1000)

	if sortBy == nil {
		fetched, err := s.CtxDB.QueriesRead.GetProductsByStatus(ctx, status.String())
		if err != nil {
			return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
		}
		for _, f := range fetched {
			products = append(products, ProductsRow(f).ToPBProduct())
		}
	} else {
		if sortBy.Field == pb.SortField_NAME {
			if sortBy.Dir == pb.SortDir_ASC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByNameAsc(ctx, status.String())
				if err != nil {
					return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}

			} else if sortBy.Dir == pb.SortDir_DESC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByNameDesc(ctx, status.String())
				if err != nil {
					return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}
			}
		} else if sortBy.Field == pb.SortField_CREATED_AT {
			if sortBy.Dir == pb.SortDir_ASC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByCreationDateDesc(ctx, status.String())
				if err != nil {
					return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}

			} else if sortBy.Dir == pb.SortDir_DESC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByCreationDateDesc(ctx, status.String())
				if err != nil {
					return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}
			}
		}
	}

	res := &pb.ProductsResponse{
		Length:   int64(len(products)),
		Products: products,
	}

	return res, nil
}

func (s *ProductServer) GetProductsListing(
	ctx context.Context,
	in *pb.GetProductsListingRequest,
) (*pb.GetProductsListingResponse, error) {
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 100
	}

	fetched, err := s.CtxDB.QueriesRead.GetProductsListing(ctx, limit)
	if err != nil {
		return nil, errs.NewGRPCError(errs.QueryFailed, err.Error())
	}

	//TODO: (Brandon)
	thumbnail := "static/images/empty.png"

	products := make([]*pb.ProductListing, 0, limit)
	for _, f := range fetched {
		moneyWithVat := utils.NewMoney(f.UnitPriceWithVat, f.UnitPriceWithVatCurrency)
		products = append(products, &pb.ProductListing{
			Id:                      serialize.EncDBID(f.ID),
			Name:                    f.Name,
			Description:             f.Description.String,
			BrandName:               f.BrandName,
			UnitPriceWithVatDisplay: moneyWithVat.Display(),
			Thumbnail:               thumbnail,
			Rating:                  0,
		})
	}

	res := &pb.GetProductsListingResponse{
		Length: int64(len(products)),
		Data:   products,
	}
	return res, nil
}
