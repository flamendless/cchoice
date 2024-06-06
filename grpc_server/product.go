package grpc_server

import (
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"context"

	"go.uber.org/zap"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
}

func (s *ProductServer) GetProduct(ctx context.Context, in *pb.ID) (*pb.Product, error) {
	logs.Log().Debug("GetProduct", zap.Int64("id", in.GetId()))
	product := &pb.Product{}
	return product, nil
}
