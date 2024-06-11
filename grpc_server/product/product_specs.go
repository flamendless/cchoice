package product

import (
	"cchoice/internal/ctx"
	"cchoice/internal/domains/grpc"
	"cchoice/internal/logs"
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
	id := in.GetId()
	logs.Log().Debug("GetProductSpecsByID", zap.Int64("id", id))

	existingProductSpecs, err := s.CtxDB.QueriesRead.GetProductSpecsByID(ctx, id)
	if err != nil {
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
	}

	return &pb.ProductSpecs{
		ID:            existingProductSpecs.ID,
		Colours:       existingProductSpecs.Colours.String,
		Sizes:         existingProductSpecs.Sizes.String,
		Segmentation:  existingProductSpecs.Segmentation.String,
		PartNumber:    existingProductSpecs.PartNumber.String,
		Power:         existingProductSpecs.Power.String,
		Capacity:      existingProductSpecs.Capacity.String,
		ScopeOfSupply: existingProductSpecs.ScopeOfSupply.String,
	}, nil
}
