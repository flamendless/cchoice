package services

import (
	"cchoice/client/common"
	pb "cchoice/proto"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"google.golang.org/grpc"
)

type AuthService struct {
	GRPCConn *grpc.ClientConn
	SM       *scs.SessionManager
}

func NewAuthService(
	grpcConn *grpc.ClientConn,
	sm *scs.SessionManager,
) AuthService {
	return AuthService{
		GRPCConn: grpcConn,
		SM:       sm,
	}
}

func (s AuthService) Authenticated(w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	tokenString := s.SM.GetString(r.Context(), "tokenString")
	if tokenString == "" {
		return &common.HandlerRes{
			Error:      errors.New("Not authenticated"),
			StatusCode: http.StatusUnauthorized,
			RedirectTo: "/auth",
		}
	}
	return nil
}

func (s AuthService) Authenticate(data *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Authenticate(ctx, data)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (s AuthService) Register(data *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Register(ctx, data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s AuthService) EnrollOTP(data *pb.EnrollOTPRequest) (*pb.EnrollOTPResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.EnrollOTP(ctx, data)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s AuthService) ValidateInitialOTP(
	data *pb.ValidateInitialOTPRequest,
) (*pb.ValidateInitialOTPResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.ValidateInitialOTP(ctx, data)
	if err != nil {
		return nil, err
	}
	return res, nil
}
