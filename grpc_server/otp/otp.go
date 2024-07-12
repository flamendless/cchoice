package otp

import (
	"cchoice/cchoice_db"
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/serialize"
	pb "cchoice/proto"
	"context"
	"database/sql"
	"fmt"
)

type OTPServer struct {
	pb.UnimplementedOTPServiceServer
	CtxDB     *ctx.Database
	Issuer    *auth.Issuer
	Validator *auth.Validator
}

func NewGRPCOTPServer(
	ctxDB *ctx.Database,
	issuer *auth.Issuer,
	validator *auth.Validator,
) *OTPServer {
	return &OTPServer{
		CtxDB:     ctxDB,
		Issuer:    issuer,
		Validator: validator,
	}
}

func (s *OTPServer) GetOTPInfo(
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

func (s *OTPServer) EnrollOTP(
	ctx context.Context,
	in *pb.EnrollOTPRequest,
) (*pb.EnrollOTPResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	otpEnabled, err := s.CtxDB.QueriesRead.GetOTPEnabledByUserID(context.Background(), userID)
	if err != nil || otpEnabled {
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
			UserID: userID,
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

func (s *OTPServer) FinishOTPEnrollment(
	ctx context.Context,
	in *pb.FinishOTPEnrollmentRequest,
) (*pb.FinishOTPEnrollmentResponse, error) {
	userID := serialize.DecDBID(in.UserId)
	otpEnabled, err := s.CtxDB.QueriesRead.GetOTPEnabledByUserID(context.Background(), userID)
	if err != nil || otpEnabled {
		return nil, errs.ERR_ALREADY_OTP_ENROLLED
	}

	resValid, err := s.ValidateOTP(
		context.Background(),
		&pb.ValidateOTPRequest{
			Passcode: in.Passcode,
			UserId:   in.UserId,
		},
	)
	if err != nil {
		return nil, err
	}
	if !resValid.Valid {
		return nil, errs.ERR_INVALID_OTP
	}

	err = s.CtxDB.Queries.FinishOTPEnrollment(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	return &pb.FinishOTPEnrollmentResponse{}, nil
}

func (s *OTPServer) GenerateOTPCode(
	ctx context.Context,
	in *pb.GenerateOTPCodeRequest,
) (*pb.GenerateOTPCodeResponse, error) {
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

		res, err := s.CtxDB.QueriesRead.GetAuthForOTPValidation(context.Background(), userID)
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
		return nil, errs.ERR_CHOOSE_VALID_OPTION
	}
	return &pb.GenerateOTPCodeResponse{}, nil
}

func (s *OTPServer) ValidateOTP(
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
		return nil, errs.ERR_INVALID_OTP
	}

	return &pb.ValidateOTPResponse{
		Valid: valid,
	}, nil
}
