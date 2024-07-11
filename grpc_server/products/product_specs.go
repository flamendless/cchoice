package products

import (
	"cchoice/internal/ctx"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"

	"go.uber.org/zap"
)

type ProductSpecsServer struct {
	pb.UnimplementedProductSpecsServiceServer
	CtxDB *ctx.Database
}

func (s *ProductSpecsServer) GetProductSpecsByID(
	ctx context.Context,
	in *pb.IDRequest,
) (*pb.ProductSpecs, error) {
	encid := in.GetId()
	logs.Log().Debug("GetProductSpecsByID", zap.String("encid", encid))

	existingProductSpecs, err := s.CtxDB.QueriesRead.GetProductSpecsByID(
		ctx,
		serialize.DecDBID(encid),
	)
	if err != nil {
		return nil, errs.NewGRPCError(errs.IDNotFound, err.Error())
	}

	return &pb.ProductSpecs{
		ID:            serialize.EncDBID(existingProductSpecs.ID),
		Colours:       existingProductSpecs.Colours.String,
		Sizes:         existingProductSpecs.Sizes.String,
		Segmentation:  existingProductSpecs.Segmentation.String,
		PartNumber:    existingProductSpecs.PartNumber.String,
		Power:         existingProductSpecs.Power.String,
		Capacity:      existingProductSpecs.Capacity.String,
		ScopeOfSupply: existingProductSpecs.ScopeOfSupply.String,
	}, nil
}
