package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	ResetTokenLength   = 32
	ResetTokenExpiry   = 15 * time.Minute
	ResetRateLimitMins = 1
)

type PasswordResetService struct {
	encoder     encode.IEncode
	dbRO        database.IService
	dbRW        database.IService
	emailRunner *jobs.EmailJobRunner
	staffLog    *StaffLogsService
}

func NewPasswordResetService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	emailRunner *jobs.EmailJobRunner,
	staffLog *StaffLogsService,
) *PasswordResetService {
	if (conf.Conf().IsProd() || conf.Conf().Test.LocalForgotPassword) && emailRunner == nil {
		panic("emailRunner is required")
	}
	if staffLog == nil {
		panic("StaffLogsService is required")
	}

	return &PasswordResetService{
		encoder:     encoder,
		dbRO:        dbRO,
		dbRW:        dbRW,
		emailRunner: emailRunner,
		staffLog:    staffLog,
	}
}

func (s *PasswordResetService) Log() {
	logs.Log().Info("[PasswordResetService]")
}

func (s *PasswordResetService) RequestReset(ctx context.Context, email string, userType enums.UserType) error {
	const logtag = "[PasswordResetService RequestReset]"

	var userID int64
	var userEmail string

	result := "request reset success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			s.encoder.Encode(userID),
			constants.ActionTrigger,
			constants.ModulePasswordReset,
			result,
			nil,
		); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}()

	switch userType {
	case enums.USER_TYPE_CUSTOMER:
		customer, err := s.dbRO.GetQueries().GetCustomerByEmail(ctx, email)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			result = err.Error()
			return err
		}
		userID = customer.ID
		userEmail = customer.Email

	case enums.USER_TYPE_STAFF:
		staff, err := s.dbRO.GetQueries().GetStaffByEmail(ctx, email)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			result = err.Error()
			return err
		}
		userID = staff.ID
		userEmail = staff.Email
	default:
		result = errs.ErrEnumInvalid.Error()
		return errs.ErrEnumInvalid
	}

	latestToken, err := s.dbRO.GetQueries().GetLatestUnusedResetToken(ctx, queries.GetLatestUnusedResetTokenParams{
		UserID:   userID,
		UserType: userType.String(),
	})
	if err == nil {
		parsedTime, parseErr := time.Parse(constants.DateTimeLayoutISO, latestToken.ExpiresAt)
		if parseErr == nil {
			if time.Since(parsedTime.Add(-ResetTokenExpiry)) < ResetRateLimitMins*time.Minute {
				result = errs.ErrPasswordResetRateLimited.Error()
				return errs.ErrPasswordResetRateLimited
			}
		}
	}

	rawToken, err := generateResetToken()
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return errs.ErrResetTokenGeneration
	}

	tokenHash := hashToken(rawToken)
	if _, err = s.dbRW.GetQueries().CreateResetToken(ctx, queries.CreateResetTokenParams{
		UserID:    userID,
		UserType:  userType.String(),
		TokenHash: tokenHash,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return errs.ErrResetTokenGeneration
	}

	resetLink := fmt.Sprintf("%s?token=%s", utils.FullURL("/auth/reset-password"), rawToken)
	if conf.Conf().IsProd() || conf.Conf().Test.LocalForgotPassword {
		if err = s.emailRunner.QueueEmailJob(ctx, jobs.EmailJobParams{
			Recipient:    userEmail,
			Subject:      "Reset Your Password - C-Choice",
			TemplateName: enums.EMAIL_TEMPLATE_PASSWORD_RESET,
			OTPCode:      resetLink,
		}); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			result = err.Error()
			return errs.ErrJobsCreateFailed
		}
	} else {
		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("reset_link", resetLink),
			zap.String("recipient", userEmail),
			zap.Stringer("usertype", userType),
		)
	}

	return nil
}

func (s *PasswordResetService) VerifyToken(ctx context.Context, token string) (*ResetContext, error) {
	const logtag = "[PasswordResetService VerifyToken]"

	tokenHash := hashToken(token)
	resetToken, err := s.dbRO.GetQueries().GetValidResetToken(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrInvalidResetToken
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return nil, errs.ErrInvalidResetToken
	}

	userType := enums.ParseUserTypeToEnum(resetToken.UserType)
	if !userType.IsValid() {
		return nil, errs.ErrInvalidResetToken
	}

	var email string
	switch userType {
	case enums.USER_TYPE_CUSTOMER:
		customer, err := s.dbRO.GetQueries().GetCustomerByID(ctx, resetToken.UserID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return nil, errs.ErrInvalidResetToken
		}
		email = customer.Email

	case enums.USER_TYPE_STAFF:
		staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, resetToken.UserID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			return nil, errs.ErrInvalidResetToken
		}
		email = staff.Email
	}

	return &ResetContext{
		UserID:   resetToken.UserID,
		UserType: userType,
		Email:    email,
	}, nil
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, token string, newPassword string) (enums.UserType, error) {
	const logtag = "[PasswordResetService ResetPassword]"

	var userID string
	result := "reset password success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			userID,
			constants.ActionTrigger,
			constants.ModulePasswordReset,
			result,
			nil,
		); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}()

	tokenHash := hashToken(token)

	resetToken, err := s.dbRO.GetQueries().GetValidResetToken(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return enums.USER_TYPE_UNDEFINED, errs.ErrInvalidResetToken
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return enums.USER_TYPE_UNDEFINED, errs.ErrInvalidResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return enums.USER_TYPE_UNDEFINED, errs.ErrPasswordResetFailed
	}

	userType := enums.ParseUserTypeToEnum(resetToken.UserType)
	switch userType {
	case enums.USER_TYPE_CUSTOMER:
		_, err = s.dbRW.GetQueries().UpdateCustomerPassword(ctx, queries.UpdateCustomerPasswordParams{
			Password: string(passwordHash),
			ID:       resetToken.UserID,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			result = err.Error()
			return enums.USER_TYPE_UNDEFINED, errs.ErrPasswordResetFailed
		}

	case enums.USER_TYPE_STAFF:
		_, err = s.dbRW.GetQueries().UpdateStaffPassword(ctx, queries.UpdateStaffPasswordParams{
			Password: string(passwordHash),
			ID:       resetToken.UserID,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			result = err.Error()
			return enums.USER_TYPE_UNDEFINED, errs.ErrPasswordResetFailed
		}
	default:
		result = errs.ErrEnumInvalid.Error()
		return enums.USER_TYPE_UNDEFINED, errs.ErrEnumInvalid
	}

	if err := s.dbRW.GetQueries().MarkResetTokenAsUsed(ctx, resetToken.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return enums.USER_TYPE_UNDEFINED, err
	}

	if err := s.dbRW.GetQueries().InvalidateUserResetTokens(ctx, queries.InvalidateUserResetTokensParams{
		UserID:   resetToken.UserID,
		UserType: resetToken.UserType,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		result = err.Error()
		return enums.USER_TYPE_UNDEFINED, err
	}

	userID = s.encoder.Encode(resetToken.UserID)
	return userType, nil
}

func generateResetToken() (string, error) {
	token := make([]byte, ResetTokenLength)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
