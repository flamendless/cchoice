package auth

import (
	"cchoice/cchoice_db"
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/alexedwards/argon2id"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	CtxDB  *ctx.Database
	Issuer *auth.Issuer
}

func (s *AuthServer) Register(
	ctx context.Context,
	in *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	err := utils.ValidateUserReg(in)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := argon2id.CreateHash(in.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %w", err)
	}

	userID, err := s.CtxDB.Queries.CreateUser(context.Background(), cchoice_db.CreateUserParams{
		FirstName:  in.FirstName,
		MiddleName: in.MiddleName,
		LastName:   in.LastName,
		Email:      in.Email,
		Password:   hashedPassword,
		MobileNo:   in.MobileNo,
		UserType:   enums.USER_TYPE_API.String(),
		Status:     enums.PRODUCT_STATUS_ACTIVE.String(),
	})
	if err != nil {
		return nil, err
	}

	err = s.CtxDB.Queries.CreateAuth(context.Background(), cchoice_db.CreateAuthParams{
		UserID:     userID,
		Token:      "",
		OtpEnabled: false,
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId: serialize.EncDBID(userID),
		Token:  "",
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

	userIDAndHashedPW, err := s.CtxDB.QueriesRead.GetUserIDAndHashedPassword(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("Invalid token in DB: %w", err)
	}

	match, err := argon2id.ComparePasswordAndHash(password, userIDAndHashedPW.Password)
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
			UserID: userIDAndHashedPW.ID,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pb.AuthenticateResponse{
		Token: tokenString,
	}, nil
}

func (s *AuthServer) EnrollOTP(
	ctx context.Context,
	in *pb.EnrollOTPRequest,
) (*pb.EnrollOTPResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	authID, err := s.CtxDB.QueriesRead.GetAuthIDByUserID(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	key, buf, err := auth.GenerateOTP(in.Issuer, in.AccountName)
	if err != nil {
		return nil, err
	}

	recoveryCodes := auth.GenerateRecoverCodes()
	secret := key.Secret()
	err = s.CtxDB.Queries.EnrollOTP(
		context.Background(),
		cchoice_db.EnrollOTPParams{
			ID: authID,
			OtpSecret: sql.NullString{
				Valid:  true,
				String: secret,
			},
			RecoveryCodes: sql.NullString{
				Valid:  true,
				String: recoveryCodes,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &pb.EnrollOTPResponse{
		Secret:        secret,
		RecoveryCodes: recoveryCodes,
		Image:         buf,
	}, nil
}

func (s *AuthServer) GetOTPCode(
	ctx context.Context,
	in *pb.GetOTPCodeRequest,
) (*pb.GetOTPCodeResponse, error) {

	var recipient cchoice_db.GetUserEMailAndMobileNoByIDRow
	var err error
	var code string

	if (in.Method == pb.OTPMethod_SMS) || (in.Method == pb.OTPMethod_EMAIL) {
		userID := serialize.DecDBID(in.UserId)
		recipient, err = s.CtxDB.QueriesRead.GetUserEMailAndMobileNoByID(
			context.Background(),
			userID,
		)
		if err != nil {
			return nil, err
		}

		res, err := s.CtxDB.QueriesRead.GetAuthIDAndSecretByUserIDAndUnvalidatedOTP(
			context.Background(),
			userID,
		)
		if err != nil {
			return nil, err
		}

		code, err = auth.GeneratePassCode(res.OtpSecret.String)
		if err != nil {
			return nil, err
		}

		fmt.Println(code, res.OtpSecret.String, recipient.MobileNo, recipient.Email)
	}

	switch in.Method {
	case pb.OTPMethod_SMS:
		//TODO: (Brandon) - send code via SMS
		break
	case pb.OTPMethod_EMAIL:
		//TODO: (Brandon) - send code via E-Mail
		break
	case pb.OTPMethod_AUTHENTICATOR:
	default:
		return nil, errors.New("Must select valid OTP Method")
	}
	return &pb.GetOTPCodeResponse{}, nil
}

func (s *AuthServer) ValidateInitialOTP(
	ctx context.Context,
	in *pb.ValidateInitialOTPRequest,
) (*pb.ValidateInitialOTPResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	res, err := s.CtxDB.QueriesRead.GetAuthIDAndSecretByUserIDAndUnvalidatedOTP(
		context.Background(),
		userID,
	)
	if err != nil {
		return nil, err
	}

	valid := auth.ValidateOTP(in.Passcode, res.OtpSecret.String)
	if !valid {
		return nil, errors.New("Invalid OTP")
	}

	err = s.CtxDB.Queries.ValidateInitialOTP(
		context.Background(),
		res.ID,
	)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateInitialOTPResponse{}, nil
}
