package grpc_server

import (
	"cchoice/internal/ctx"
	"cchoice/internal/domains/grpc"
	"cchoice/internal/logs"
	"cchoice/internal/models"
	pb "cchoice/proto"
	"context"

	"go.uber.org/zap"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	ctxDB *ctx.Database
}

func (s *ProductServer) GetProductByID(ctx context.Context, in *pb.ID) (*pb.Product, error) {
	logs.Log().Debug("GetProductByID", zap.Int64("id", in.GetId()))

	existingProduct, err := s.ctxDB.QueriesRead.GetProductByID(ctx, in.GetId())
	if err != nil {
		return nil, grpc.NewGRPCError(grpc.IDNotFound, err.Error())
	}

	res := models.DBRowToProductPB(&existingProduct)
	return res, nil
}
