package auth

import (
	"cchoice/cchoice_db"
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	CtxDB     *ctx.Database
	Issuer    *auth.Issuer
	Validator *auth.Validator
}

func NewGRPCAuthServer(
	ctxDB *ctx.Database,
	issuer *auth.Issuer,
	validator *auth.Validator,
) *AuthServer {
	return &AuthServer{
		CtxDB:     ctxDB,
		Issuer:    issuer,
		Validator: validator,
	}
}

func (s *AuthServer) ValidateToken(
	ctx context.Context,
	in *pb.ValidateTokenRequest,
) (*pb.ValidateTokenResponse, error) {
	expectedAUD := enums.ParseAudEnum(in.Aud)
	res, err := s.Validator.GetToken(expectedAUD, in.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid auth token: %v", err)
	}
	return &pb.ValidateTokenResponse{
		Success: true,
		UserId:  res.UserID,
	}, nil
}

func (s *AuthServer) Authenticate(
	ctx context.Context,
	in *pb.AuthenticateRequest,
) (*pb.AuthenticateResponse, error) {
	username := in.GetUsername()
	password := in.GetPassword()

	errUsername := utils.ValidateUsername(username)
	if errUsername != nil {
		return nil, errUsername
	}

	errPW := utils.ValidatePW(password)
	if errPW != nil {
		return nil, errPW
	}

	resUser, err := s.CtxDB.QueriesRead.GetUserForAuth(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	match, err := argon2id.ComparePasswordAndHash(password, resUser.Password)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	if !match {
		return nil, fmt.Errorf("Invalid credentials")
	}

	tokenString, err := s.Issuer.IssueToken(enums.AUD_API, username)
	if err != nil {
		return nil, err
	}

	err = s.CtxDB.Queries.UpdateAuthTokenByUserID(
		context.Background(),
		cchoice_db.UpdateAuthTokenByUserIDParams{
			Token:  tokenString,
			UserID: resUser.ID,
		},
	)
	if err != nil {
		return nil, err
	}

	// resOTP, err := s.CtxDB.QueriesRead.GetAuthOTP(context.Background(), resUser.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// if resOTP.OtpEnabled {
	// 	eOTPStatus := enums.ParseOTPStatusEnum(resOTP.OtpStatus)
	// 	if eOTPStatus == enums.OTP_STATUS_ENROLLED || eOTPStatus == enums.OTP_STATUS_SENT_CODE {
	// 		err = s.CtxDB.Queries.NeedOTP(context.Background(), resUser.ID)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		return &pb.AuthenticateResponse{
	// 			Token:   tokenString,
	// 			NeedOtp: true,
	// 		}, nil
	// 	}
	// }

	return &pb.AuthenticateResponse{
		Token: tokenString,
	}, nil
}
