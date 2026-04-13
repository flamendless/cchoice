package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"

	"go.uber.org/zap"
)

const (
	OTPCodeLength    = 6
	OTPExpiryMins    = 5
	OTPRateLimitMins = 1
)

type CustomerOTPService struct {
	encoder     encode.IEncode
	dbRO        database.IService
	dbRW        database.IService
	mailService mail.IMailService
	emailRunner *jobs.EmailJobRunner
}

func NewCustomerOTPService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	mailService mail.IMailService,
	emailRunner *jobs.EmailJobRunner,
) *CustomerOTPService {
	if (conf.Conf().IsProd() || conf.Conf().Test.LocalOTP) && emailRunner == nil {
		panic("emailRunner is required")
	}
	return &CustomerOTPService{
		encoder:     encoder,
		dbRO:        dbRO,
		dbRW:        dbRW,
		mailService: mailService,
		emailRunner: emailRunner,
	}
}

func (s *CustomerOTPService) GenerateAndSendVerificationCode(ctx context.Context, params GenerateOTPParams) error {
	const logtag = "[CustomerOTPService GenerateAndSendVerificationCode]"

	customerID := s.encoder.Decode(params.CustomerID)
	if customerID == encode.INVALID {
		return errs.ErrDecode
	}

	latestOTP, err := s.dbRO.GetQueries().GetLatestUnusedOTPCode(ctx, customerID)
	if err == nil {
		parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", latestOTP.ExpiresAt)
		if parseErr == nil {
			if time.Since(parsedTime) < OTPRateLimitMins*time.Minute {
				return errs.ErrOTPRateLimited
			}
		}
	}

	otpCode, err := generateRandomOTP()
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errs.ErrOTPGenerationFailed
	}

	_, err = s.dbRW.GetQueries().CreateOTPCode(ctx, queries.CreateOTPCodeParams{
		CustomerID: customerID,
		OtpCode:    otpCode,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errs.ErrOTPCreationFailed
	}

	if conf.Conf().IsProd() || conf.Conf().Test.LocalOTP {
		if err = s.emailRunner.QueueEmailJob(ctx, jobs.EmailJobParams{
			Recipient:    params.Email,
			Subject:      "Verify Your Email - C-Choice",
			TemplateName: enums.EMAIL_TEMPLATE_CUSTOMER_VERIFICATION,
			EMail:        params.Email,
			OTPCode:      otpCode,
		}); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return errs.ErrJobsCreateFailed
		}
	} else {
		logs.Log().Info("[DEBUG]", zap.String("otp", otpCode))
	}

	return nil
}

func (s *CustomerOTPService) VerifyCode(ctx context.Context, customerID string, code string) (bool, error) {
	const logtag = "[CustomerOTPService VerifyCode]"

	dbCustomerID := s.encoder.Decode(customerID)
	if dbCustomerID == encode.INVALID {
		return false, errs.ErrDecode
	}

	validOTP, err := s.dbRO.GetQueries().GetValidOTPCode(ctx, queries.GetValidOTPCodeParams{
		CustomerID: dbCustomerID,
		OtpCode:    code,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return false, errs.ErrInvalidOTP
	}

	err = s.dbRW.GetQueries().MarkOTPAsUsed(ctx, validOTP.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return false, errs.ErrOTPUpdateFailed
	}

	_, err = s.dbRW.GetQueries().UpdateCustomerStatusToVerified(ctx, dbCustomerID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return false, errs.ErrCustomerStatusUpdateFailed
	}

	return true, nil
}

func generateRandomOTP() (string, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(OTPCodeLength), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return fmt.Sprintf("%06d", n), nil
}

func (s *CustomerOTPService) Log() {
	logs.Log().Info("[CustomerOTPService] Loaded")
}

var _ IService = (*CustomerOTPService)(nil)
