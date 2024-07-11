package services

import (
	"cchoice/client/common"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	pb "cchoice/proto"
	"context"
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

func (s AuthService) Authenticated(aud enums.AudKind, w http.ResponseWriter, r *http.Request) *common.HandlerRes {
	tokenString := s.SM.GetString(r.Context(), "tokenString")
	if tokenString == "" {
		return &common.HandlerRes{
			Error:      errs.ERR_NO_AUTH,
			StatusCode: http.StatusUnauthorized,
			RedirectTo: "/auth",
		}
	}

	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Aud:   aud.String(),
		Token: tokenString,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_NO_AUTH,
			StatusCode: http.StatusUnauthorized,
			RedirectTo: "/auth",
		}
	}
	return nil
}

func (s AuthService) Authenticate(data *common.AuthAuthenticateRequest) (string, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Authenticate(ctx, &pb.AuthenticateRequest{
		Username: data.Username,
		Password: data.Password,
	})
	if err != nil {
		return "", err
	}
	return res.Token, err
}

func (s AuthService) Register(data *common.AuthRegisterRequest) (string, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Register(ctx, &pb.RegisterRequest{
		FirstName:       data.FirstName,
		MiddleName:      data.MiddleName,
		LastName:        data.LastName,
		Email:           data.Email,
		Password:        data.Password,
		ConfirmPassword: data.ConfirmPassword,
		MobileNo:        data.MobileNo,
	})
	if err != nil {
		return "", err
	}
	return res.UserId, nil
}

func (s AuthService) EnrollOTP(data *common.AuthEnrollOTPRequest) (*common.AuthEnrollOTPResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.EnrollOTP(ctx, &pb.EnrollOTPRequest{
		UserId:      data.UserID,
		Issuer:      data.Issuer,
		AccountName: data.AccountName,
	})
	if err != nil {
		return nil, err
	}
	return &common.AuthEnrollOTPResponse{
		Secret:        res.Secret,
		RecoveryCodes: res.RecoveryCodes,
		Image:         res.Image,
	}, nil
}

func (s AuthService) GetOTPCode(data *common.AuthGetOTPCodeRequest) error {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.GetOTPCode(ctx, &pb.GetOTPCodeRequest{
		UserId: data.UserID,
		Method: enums.StringToPBEnum(
			data.Method,
			pb.OTPMethod_OTPMethod_value,
			pb.OTPMethod_UNDEFINED,
		),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s AuthService) ValidateInitialOTP(data *common.AuthValidateInitialOTP) error {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.ValidateInitialOTP(ctx, &pb.ValidateInitialOTPRequest{
		UserId:   data.UserID,
		Passcode: data.Passcode,
	})
	if err != nil {
		return err
	}
	return nil
}
