package auth

import (
	"cchoice/cchoice_db"
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"
	pb "cchoice/proto"
	"context"
	"database/sql"
	"errors"
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
		Status:     pb.UserStatus_ACTIVE.String(),
	})
	if err != nil {
		return nil, err
	}

	err = s.CtxDB.Queries.CreateInitialAuth(context.Background(), userID)
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

	resOTP, err := s.CtxDB.QueriesRead.GetAuthOTP(context.Background(), resUser.ID)
	if err != nil {
		return nil, err
	}
	if resOTP.OtpEnabled {
		eOTPStatus := enums.ParseOTPStatusEnum(resOTP.OtpStatus)
		if eOTPStatus == enums.OTP_STATUS_ENROLLED || eOTPStatus == enums.OTP_STATUS_SENT_CODE {
			err = s.CtxDB.Queries.NeedOTP(context.Background(), resUser.ID)
			if err != nil {
				return nil, err
			}
			return &pb.AuthenticateResponse{
				Token:   tokenString,
				NeedOtp: true,
			}, nil
		}
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
	authID, err := s.CtxDB.QueriesRead.GetAuthForEnrollmentByUserID(context.Background(), userID)
	if err != nil {
		return nil, errs.ERR_ALREADY_OTP_ENROLLED
	}

	key, buf, err := auth.GenerateOTP(in.Issuer, in.AccountName)
	if err != nil {
		return nil, err
	}

	recoveryCodes := auth.GenerateRecoveryCodes()
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

		res, err := s.CtxDB.QueriesRead.GetAuthForOTPValidation(
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

func (s *AuthServer) FinishOTPEnrollment(
	ctx context.Context,
	in *pb.FinishOTPEnrollmentRequest,
) (*pb.FinishOTPEnrollmentResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	res, err := s.CtxDB.QueriesRead.GetAuthForOTPValidation(
		context.Background(),
		userID,
	)
	if err != nil {
		return nil, errs.ERR_ALREADY_OTP_ENROLLED
	}

	resValid, err := s.ValidateOTP(context.Background(), &pb.ValidateOTPRequest{
		Passcode: in.Passcode,
		UserId:   in.UserId,
	})
	if err != nil {
		return nil, err
	}
	if !resValid.Valid {
		return nil, errors.New("Invalid OTP")
	}

	err = s.CtxDB.Queries.FinishOTPEnrollment(
		context.Background(),
		res.ID,
	)
	if err != nil {
		return nil, err
	}

	return &pb.FinishOTPEnrollmentResponse{}, nil
}

func (s *AuthServer) GetOTPInfo(
	ctx context.Context,
	in *pb.GetOTPInfoRequest,
) (*pb.GetOTPInfoResponse, error) {
	info, err := s.CtxDB.QueriesRead.GetUserEMailAndMobileNoByID(
		context.Background(),
		serialize.DecDBID(in.UserId),
	)
	if err != nil {
		return nil, err
	}

	var recipient string

	eOTPMethod := enums.StringToPBEnum(
		in.OtpMethod,
		pb.OTPMethod_OTPMethod_value,
		pb.OTPMethod_UNDEFINED,
	)

	switch eOTPMethod {
	case pb.OTPMethod_UNDEFINED:
		return nil, errs.ERR_CHOOSE_VALID_OPTION
	case pb.OTPMethod_AUTHENTICATOR:
		recipient = "Authenticator"
	case pb.OTPMethod_SMS:
		recipient = info.MobileNo
	case pb.OTPMethod_EMAIL:
		recipient = info.Email
	}

	return &pb.GetOTPInfoResponse{
		Recipient: recipient,
	}, nil
}

func (s *AuthServer) ValidateOTP(
	ctx context.Context,
	in *pb.ValidateOTPRequest,
) (*pb.ValidateOTPResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	res, err := s.CtxDB.QueriesRead.GetAuthForOTPValidation(
		context.Background(),
		userID,
	)
	if err != nil {
		return nil, errs.ERR_ALREADY_OTP_ENROLLED
	}

	valid := auth.ValidateOTP(in.Passcode, res.OtpSecret.String)
	if !valid {
		return nil, errors.New("Invalid OTP")
	}

	err = s.CtxDB.Queries.SetOTPStatusValidByUserID(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateOTPResponse{
		Valid: valid,
	}, nil
}
