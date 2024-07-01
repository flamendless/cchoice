package services

import (
	pb "cchoice/proto"
	"context"
	"time"

	"google.golang.org/grpc"
)

type AuthService struct {
	GRPCConn *grpc.ClientConn
}

func NewAuthService(grpcConn *grpc.ClientConn) AuthService {
	return AuthService{
		GRPCConn: grpcConn,
	}
}

func (s AuthService) Authenticate(
	username string,
	password string,
) (*pb.AuthLoginResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.Authenticate(ctx, &pb.AuthLoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return res, err
}
