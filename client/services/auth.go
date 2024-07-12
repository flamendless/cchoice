package services

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/internal/auth"
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

func (s AuthService) Authenticated(aud enums.AudKind, w http.ResponseWriter, r *http.Request) (*common.HandlerRes, auth.ValidToken) {
	tokenString := s.SM.GetString(r.Context(), "tokenString")
	token := auth.ValidToken{}
	if tokenString == "" {
		return &common.HandlerRes{
			Error:      errs.ERR_NO_AUTH,
			StatusCode: http.StatusUnauthorized,
			RedirectTo: "/auth",
		}, token
	}

	needOTP := s.SM.GetBool(r.Context(), "needOTP")
	if needOTP {
		return &common.HandlerRes{
			Component: components.CenterCard(
				components.OTPView(false),
			),
			ReplaceURL: "/otp",
		}, token
	}

	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Aud:   aud.String(),
		Token: tokenString,
	})
	if err != nil {
		return &common.HandlerRes{
			Error:      errs.ERR_NO_AUTH,
			StatusCode: http.StatusUnauthorized,
			RedirectTo: "/auth",
		}, token
	}

	token.UserID = res.UserId
	return nil, token
}

func (s AuthService) Authenticate(data *common.AuthAuthenticateRequest) (*common.AuthAuthenticateResponse, error) {
	client := pb.NewAuthServiceClient(s.GRPCConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Authenticate(ctx, &pb.AuthenticateRequest{
		Username: data.Username,
		Password: data.Password,
	})
	if err != nil {
		return nil, err
	}
	return &common.AuthAuthenticateResponse{
		Token:   res.Token,
		NeedOTP: res.NeedOtp,
	}, nil
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

func (s AuthService) FinishOTPEnrollment(data *common.AuthFinishOTPEnrollment) error {
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

func (s AuthService) ValidateToken(data *common.AuthValidateTokenRequest) (*common.AuthValidateTokenResponse, error) {
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

func (s AuthService) GetOTPInfo(data *common.AuthGetOTPInfoRequest) (*common.AuthGetOTPInfoResponse, error) {
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

func (s AuthService) ValidateOTP(data *common.AuthValidateOTPRequest) error {
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
