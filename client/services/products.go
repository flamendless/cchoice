package services

import (
	pb "cchoice/proto"
	"context"
	"time"

	"google.golang.org/grpc"
)

type ProductService struct {
	GRPCConn *grpc.ClientConn
}

func NewProductService(grpcConn *grpc.ClientConn) ProductService {
	return ProductService{
		GRPCConn: grpcConn,
	}
}

func (s *ProductService) GetProductsWithSorting(
	sortField pb.SortField,
	sortDir pb.SortDir,
) (*pb.ProductsResponse, error) {
	client := pb.NewProductServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.ListProductsByProductStatus(
		ctx,
		&pb.ProductStatusRequest{
			Status: pb.ProductStatus_ACTIVE,
			SortBy: &pb.SortBy{
				Field: sortField,
				Dir:   sortDir,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}
