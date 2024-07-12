package services

import (
	"cchoice/client/common"
	"cchoice/internal/enums"
	pb "cchoice/proto"
	"context"
	"time"

	"github.com/alexedwards/scs/v2"
	"google.golang.org/grpc"
)

type OTPService struct {
	GRPCConn *grpc.ClientConn
	SM       *scs.SessionManager
}

func NewOTPService(
	grpcConn *grpc.ClientConn,
	sm *scs.SessionManager,
) AuthService {
	return AuthService{
		GRPCConn: grpcConn,
		SM:       sm,
	}
}

func (s OTPService) EnrollOTP(data *common.AuthEnrollOTPRequest) (*common.AuthEnrollOTPResponse, error) {
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

func (s OTPService) GetOTPCode(data *common.AuthGetOTPCodeRequest) error {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.GetOTPCode(ctx, &pb.GetOTPCodeRequest{
		UserId: data.UserID,
		Method: enums.StringToPBEnum(
			data.OTPMethod,
			pb.OTPMethod_OTPMethod_value,
			pb.OTPMethod_UNDEFINED,
		),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s OTPService) FinishOTPEnrollment(data *common.AuthFinishOTPEnrollment) error {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.FinishOTPEnrollment(ctx, &pb.FinishOTPEnrollmentRequest{
		UserId:   data.UserID,
		Passcode: data.Passcode,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s OTPService) ValidateToken(data *common.AuthValidateTokenRequest) (*common.AuthValidateTokenResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: data.Token,
		Aud:   data.AUD,
	})
	if err != nil {
		return nil, err
	}
	return &common.AuthValidateTokenResponse{
		UserID: res.UserId,
	}, nil
}

func (s OTPService) GetOTPInfo(data *common.AuthGetOTPInfoRequest) (*common.AuthGetOTPInfoResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.GetOTPInfo(ctx, &pb.GetOTPInfoRequest{
		UserId:    data.UserID,
		OtpMethod: data.OTPMethod,
	})
	if err != nil {
		return nil, err
	}
	return &common.AuthGetOTPInfoResponse{
		Recipient: res.Recipient,
	}, nil
}

func (s OTPService) ValidateOTP(data *common.AuthValidateOTPRequest) error {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.ValidateOTP(ctx, &pb.ValidateOTPRequest{
		Passcode: data.Passcode,
		UserId:   data.UserID,
	})
	if err != nil {
		return err
	}
	return  nil
}
