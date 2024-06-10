package product

import (
	"cchoice/cchoice_db"
	"cchoice/internal/ctx"
	"cchoice/internal/domains/grpc"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"context"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	CtxDB *ctx.Database
}

func productFromRow(row *cchoice_db.GetProductByIDRow) *pb.Product {
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
		ID:          row.ID,
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

func (s *ProductServer) GetByID(ctx context.Context, in *pb.ID) (*pb.Product, error) {
	id := in.GetId()
	logs.Log().Debug("GetByID", zap.Int64("id", id))

	existingProduct, err := s.CtxDB.QueriesRead.GetProductByID(ctx, id)
	if err != nil {
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
	}

	res := productFromRow(&existingProduct)
	return res, nil
}
