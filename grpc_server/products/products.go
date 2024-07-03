package products

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/domains/grpc"
	"cchoice/internal/enums"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	CtxDB *ctx.Database
}

type ProductsRow cchoice_db.GetProductsRow

func (row ProductsRow) ToPBProduct() *pb.Product {
	unitPriceWithoutVat := decimal.NewFromInt(row.UnitPriceWithoutVat / 100)
	unitPriceWithVat := decimal.NewFromInt(row.UnitPriceWithVat / 100)
	moneyWithoutVat := money.New(
		unitPriceWithoutVat.CoefficientInt64(),
		row.UnitPriceWithoutVatCurrency,
	)
	moneyWithVat := money.New(
		unitPriceWithVat.CoefficientInt64(),
		row.UnitPriceWithVatCurrency,
	)

	return &pb.Product{
		ID:          serialize.EncDBID(row.ID),
		Serial:      row.Serial,
		Name:        row.Name,
		Description: row.Description.String,
		Brand:       row.Brand,
		Status:      enums.ParseProductStatusEnumPB(row.Status),
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
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
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
			return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
		}
		for _, f := range fetched {
			products = append(products, ProductsRow(f).ToPBProduct())
		}
	} else {
		if sortBy.Field == pb.SortField_NAME {
			if sortBy.Dir == pb.SortDir_ASC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByNameAsc(ctx, status.String())
				if err != nil {
					return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}

			} else if sortBy.Dir == pb.SortDir_DESC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByNameDesc(ctx, status.String())
				if err != nil {
					return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}
			}
		} else if sortBy.Field == pb.SortField_CREATED_AT {
			if sortBy.Dir == pb.SortDir_ASC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByCreationDateDesc(ctx, status.String())
				if err != nil {
					return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
				}
				for _, f := range fetched {
					products = append(products, ProductsRow(f).ToPBProduct())
				}

			} else if sortBy.Dir == pb.SortDir_DESC {
				fetched, err := s.CtxDB.QueriesRead.GetProductsByStatusSortByCreationDateDesc(ctx, status.String())
				if err != nil {
					return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
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
