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

func (s AuthService) Authenticated(r *http.Request) *common.HandlerRes {
	tokenString := s.SM.GetString(r.Context(), "tokenString")
	if tokenString == "" {
		return &common.HandlerRes{
			Error:      errors.New("Not authenticated"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	return nil
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
