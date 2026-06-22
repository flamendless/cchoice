package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
)

const MemoEmailCooldown = 24 * time.Hour

type MemoService struct {
	encoder     encode.IEncode
	dbRO        database.IService
	dbRW        database.IService
	staffLog    *StaffLogsService
	emailRunner *jobs.EmailJobRunner
}

func NewMemoService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
	emailRunner *jobs.EmailJobRunner,
) *MemoService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	if (conf.Conf().IsProd() || (conf.Conf().IsLocal() && conf.Conf().Test.LocalMemoEmailSend)) && emailRunner == nil {
		panic("emailRunner is required")
	}
	return &MemoService{
		encoder:     encoder,
		dbRO:        dbRO,
		dbRW:        dbRW,
		staffLog:    staffLog,
		emailRunner: emailRunner,
	}
}

func (s *MemoService) GetMemoByID(ctx context.Context, memoID string) (*Memo, error) {
	id := s.encoder.Decode(memoID)
	if id == encode.INVALID {
		return nil, errs.ErrDecode
	}

	row, err := s.dbRO.GetQueries().GetMemoByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(errs.ErrMemo, err)
	}

	memo := s.mapRowToMemo(row.TblMemo)
	return &memo, nil
}

func (s *MemoService) GetAllForAdmin(ctx context.Context) ([]MemoListItem, error) {
	rows, err := s.dbRO.GetQueries().GetAllMemos(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrMemo, err)
	}

	result := make([]MemoListItem, 0, len(rows))
	for _, row := range rows {
		memo := s.mapRowToMemo(row.TblMemo)
		result = append(result, MemoListItem{
			Memo: memo,
			CreatedByName: utils.BuildFullName(
				row.CreatorFirstName,
				row.CreatorMiddleName.String,
				row.CreatorLastName,
			),
			CreatorPosition: row.CreatorPosition,
		})
	}
	return result, nil
}

func (s *MemoService) SendMemoEmails(ctx context.Context, actorStaffID, memoID string, isSuperuser bool) error {
	const logtag = "[MemoService SendMemoEmails]"

	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			actorStaffID,
			constants.ActionTrigger,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
	}()

	decodedMemoID := s.encoder.Decode(memoID)
	if decodedMemoID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	decodedActorID := s.encoder.Decode(actorStaffID)
	if decodedActorID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	memo, err := s.dbRO.GetQueries().GetMemoByID(ctx, decodedMemoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = errs.ErrMemoNotFound.Error()
			return errs.ErrMemoNotFound
		}
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	if memo.TblMemo.Status != enums.MEMO_STATUS_PUBLISHED.String() {
		result = errs.ErrMemoNotPublished.Error()
		return errs.ErrMemoNotPublished
	}

	if !isSuperuser && memo.TblMemo.CreatedBy != decodedActorID {
		result = errs.ErrMemoSendNotAllowed.Error()
		return errs.ErrMemoSendNotAllowed
	}

	if err := s.checkMemoEmailCooldown(memo.TblMemo.EmailsSentAt); err != nil {
		result = err.Error()
		return err
	}

	recipients, err := s.dbRO.GetQueries().GetMemoRecipientEmails(ctx, decodedMemoID)
	if err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}
	if len(recipients) == 0 {
		result = errs.ErrMemoNoRecipientEmails.Error()
		return errs.ErrMemoNoRecipientEmails
	}

	if err := s.dbRW.GetQueries().UpdateMemoEmailsSentAt(ctx, decodedMemoID); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	if conf.Conf().IsProd() || (conf.Conf().IsLocal() && conf.Conf().Test.LocalMemoEmailSend) {
		memoIDCopy := decodedMemoID
		subject := fmt.Sprintf("New Memo: %s - C-Choice", memo.TblMemo.Title)
		cc := conf.Conf().MailerooConfig.CC
		for _, recipient := range recipients {
			if err := s.emailRunner.QueueEmailJob(ctx, jobs.EmailJobParams{
				MemoID:       &memoIDCopy,
				Recipient:    recipient.Email,
				CC:           cc,
				Subject:      subject,
				TemplateName: enums.EMAIL_TEMPLATE_MEMO_NOTIFICATION,
			}); err != nil {
				result = err.Error()
				logs.LogCtx(ctx).Error(logtag, zap.Error(err))
				return err
			}
		}
	} else {
		logs.LogCtx(ctx).Info(logtag, zap.String("result", "skipped (non-prod)"), zap.Int("recipient_count", len(recipients)))
	}

	result = fmt.Sprintf("success. memo ID '%s', recipients %d", memoID, len(recipients))
	return nil
}

func (s *MemoService) checkMemoEmailCooldown(emailsSentAt string) error {
	if emailsSentAt == "" || strings.HasPrefix(emailsSentAt, "1970-01-01") {
		return nil
	}

	sentAt, err := time.Parse(constants.DateTimeLayoutISO, emailsSentAt)
	if err != nil {
		sentAt, err = time.Parse(constants.DateTimeLayoutTZISO, emailsSentAt)
		if err != nil {
			return nil
		}
	}

	if time.Since(sentAt) < MemoEmailCooldown {
		return errs.ErrMemoEmailRateLimited
	}
	return nil
}

func (s *MemoService) GetRecipientStaffIDs(ctx context.Context, memoID string) ([]string, error) {
	id := s.encoder.Decode(memoID)
	if id == encode.INVALID {
		return nil, errs.ErrDecode
	}

	staffIDs, err := s.dbRO.GetQueries().GetMemoRecipientStaffIDs(ctx, id)
	if err != nil {
		return nil, errors.Join(errs.ErrMemo, err)
	}

	result := make([]string, 0, len(staffIDs))
	for _, staffID := range staffIDs {
		result = append(result, s.encoder.Encode(staffID))
	}
	return result, nil
}

func (s *MemoService) GetRecipientsWithActions(ctx context.Context, memoID string) ([]MemoRecipientRow, error) {
	id := s.encoder.Decode(memoID)
	if id == encode.INVALID {
		return nil, errs.ErrDecode
	}

	rows, err := s.dbRO.GetQueries().GetMemoRecipientsWithActions(ctx, id)
	if err != nil {
		return nil, errors.Join(errs.ErrMemo, err)
	}

	result := make([]MemoRecipientRow, 0, len(rows))
	for _, row := range rows {
		actionStatus := enums.MEMO_STAFF_ACTION_STATUS_UNDEFINED
		if row.ActionStatus.Valid {
			actionStatus = enums.ParseMemoStaffActionStatusToEnum(row.ActionStatus.String)
		}

		result = append(result, MemoRecipientRow{
			StaffID: s.encoder.Encode(row.StaffID),
			StaffName: utils.BuildFullName(
				row.FirstName,
				row.MiddleName.String,
				row.LastName,
			),
			ActionStatus: actionStatus,
			RejectReason: row.RejectReason.String,
			AcceptedAt:   row.AcceptedAt.String,
			RejectedAt:   row.RejectedAt.String,
		})
	}
	return result, nil
}

func (s *MemoService) GetPendingForStaff(ctx context.Context, staffID string) ([]StaffPendingMemo, error) {
	decodedStaffID := s.encoder.Decode(staffID)
	if decodedStaffID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	rows, err := s.dbRO.GetQueries().GetPendingMemosForStaff(ctx, queries.GetPendingMemosForStaffParams{
		StaffID:   decodedStaffID,
		StaffID_2: decodedStaffID,
	})
	if err != nil {
		return nil, errors.Join(errs.ErrMemo, err)
	}

	result := make([]StaffPendingMemo, 0, len(rows))
	for _, row := range rows {
		result = append(result, StaffPendingMemo{
			ID:      s.encoder.Encode(row.TblMemo.ID),
			Title:   row.TblMemo.Title,
			Message: row.TblMemo.Message,
			FileURL: row.TblMemo.FileUrl.String,
		})
	}
	return result, nil
}

func (s *MemoService) CreateMemo(
	ctx context.Context,
	actorStaffID string,
	title string,
	message string,
	fileURL string,
	status enums.MemoStatus,
	startDate time.Time,
	endDate time.Time,
	recipientStaffIDs []string,
) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			actorStaffID,
			constants.ActionCreate,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if err := ValidateMemoDates(startDate, endDate); err != nil {
		result = err.Error()
		return "", err
	}

	if len(recipientStaffIDs) == 0 {
		result = errs.ErrMemoRecipientsRequired.Error()
		return "", errs.ErrMemoRecipientsRequired
	}

	createdBy := s.encoder.Decode(actorStaffID)
	if createdBy == encode.INVALID {
		result = errs.ErrDecode.Error()
		return "", errs.ErrDecode
	}

	decodedRecipients, err := s.decodeStaffIDs(recipientStaffIDs)
	if err != nil {
		result = err.Error()
		return "", err
	}

	id, err := s.dbRW.GetQueries().CreateMemo(ctx, queries.CreateMemoParams{
		Title:     title,
		Message:   message,
		FileUrl:   sql.NullString{Valid: fileURL != "", String: fileURL},
		Status:    status.String(),
		StartDate: startDate.Format(constants.DateLayoutISO),
		EndDate:   endDate.Format(constants.DateLayoutISO),
		CreatedBy: createdBy,
	})
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrMemo, err)
	}

	if err := s.replaceRecipients(ctx, id, decodedRecipients); err != nil {
		result = err.Error()
		return "", err
	}

	memoID := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", memoID)
	return memoID, nil
}

func (s *MemoService) UpdateMemo(
	ctx context.Context,
	actorStaffID string,
	memoID string,
	title string,
	message string,
	fileURL string,
	status enums.MemoStatus,
	startDate time.Time,
	endDate time.Time,
	recipientStaffIDs []string,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			actorStaffID,
			constants.ActionUpdate,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if err := ValidateMemoDates(startDate, endDate); err != nil {
		result = err.Error()
		return err
	}

	if len(recipientStaffIDs) == 0 {
		result = errs.ErrMemoRecipientsRequired.Error()
		return errs.ErrMemoRecipientsRequired
	}

	id := s.encoder.Decode(memoID)
	if id == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	memo, err := s.GetMemoByID(ctx, memoID)
	if err != nil {
		result = err.Error()
		return err
	}
	if memo == nil {
		result = errs.ErrMemoNotFound.Error()
		return errs.ErrMemoNotFound
	}

	decodedRecipients, err := s.decodeStaffIDs(recipientStaffIDs)
	if err != nil {
		result = err.Error()
		return err
	}

	if err := s.dbRW.GetQueries().UpdateMemo(ctx, queries.UpdateMemoParams{
		Title:     title,
		Message:   message,
		FileUrl:   sql.NullString{Valid: fileURL != "", String: fileURL},
		Status:    status.String(),
		StartDate: startDate.Format(constants.DateLayoutISO),
		EndDate:   endDate.Format(constants.DateLayoutISO),
		ID:        id,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	if err := s.replaceRecipients(ctx, id, decodedRecipients); err != nil {
		result = err.Error()
		return err
	}

	result = fmt.Sprintf("success. ID '%s'", memoID)
	return nil
}

func (s *MemoService) SoftDeleteMemo(ctx context.Context, actorStaffID string, memoID string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			actorStaffID,
			constants.ActionDelete,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := s.encoder.Decode(memoID)
	if id == encode.INVALID {
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().SoftDeleteMemo(ctx, id); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	result = fmt.Sprintf("success. ID '%s'", memoID)
	return nil
}

func (s *MemoService) AcceptMemo(ctx context.Context, staffID string, memoID string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionAccept,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if err := s.validateStaffAction(ctx, staffID, memoID); err != nil {
		result = err.Error()
		return err
	}

	decodedStaffID := s.encoder.Decode(staffID)
	decodedMemoID := s.encoder.Decode(memoID)

	if _, err := s.dbRW.GetQueries().CreateMemoStaffActionAccept(ctx, queries.CreateMemoStaffActionAcceptParams{
		MemoID:  decodedMemoID,
		StaffID: decodedStaffID,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	result = fmt.Sprintf("success. memo ID '%s'", memoID)
	return nil
}

func (s *MemoService) RejectMemo(ctx context.Context, staffID string, memoID string, reason string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionReject,
			constants.ModuleMemos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if reason == "" {
		result = errs.ErrMemoRejectReasonRequired.Error()
		return errs.ErrMemoRejectReasonRequired
	}

	if err := s.validateStaffAction(ctx, staffID, memoID); err != nil {
		result = err.Error()
		return err
	}

	decodedStaffID := s.encoder.Decode(staffID)
	decodedMemoID := s.encoder.Decode(memoID)

	if _, err := s.dbRW.GetQueries().CreateMemoStaffActionReject(ctx, queries.CreateMemoStaffActionRejectParams{
		MemoID:       decodedMemoID,
		StaffID:      decodedStaffID,
		RejectReason: sql.NullString{Valid: true, String: reason},
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrMemo, err)
	}

	result = fmt.Sprintf("success. memo ID '%s'", memoID)
	return nil
}

func (s *MemoService) validateStaffAction(ctx context.Context, staffID string, memoID string) error {
	decodedStaffID := s.encoder.Decode(staffID)
	if decodedStaffID == encode.INVALID {
		return errs.ErrDecode
	}

	decodedMemoID := s.encoder.Decode(memoID)
	if decodedMemoID == encode.INVALID {
		return errs.ErrDecode
	}

	count, err := s.dbRO.GetQueries().IsMemoRecipient(ctx, queries.IsMemoRecipientParams{
		MemoID:  decodedMemoID,
		StaffID: decodedStaffID,
	})
	if err != nil {
		return errors.Join(errs.ErrMemo, err)
	}
	if count == 0 {
		return errs.ErrMemoNotFound
	}

	_, err = s.dbRO.GetQueries().GetMemoStaffAction(ctx, queries.GetMemoStaffActionParams{
		MemoID:  decodedMemoID,
		StaffID: decodedStaffID,
	})
	if err == nil {
		return errs.ErrMemoAlreadyAcknowledged
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return errors.Join(errs.ErrMemo, err)
	}

	pending, err := s.GetPendingForStaff(ctx, staffID)
	if err != nil {
		return err
	}

	for _, m := range pending {
		if m.ID == memoID {
			return nil
		}
	}

	return errs.ErrMemoNotFound
}

func (s *MemoService) replaceRecipients(ctx context.Context, memoID int64, staffIDs []int64) error {
	if err := s.dbRW.GetQueries().DeleteMemoRecipientsByMemoID(ctx, memoID); err != nil {
		return errors.Join(errs.ErrMemo, err)
	}

	for _, staffID := range staffIDs {
		if _, err := s.dbRW.GetQueries().CreateMemoRecipient(ctx, queries.CreateMemoRecipientParams{
			MemoID:  memoID,
			StaffID: staffID,
		}); err != nil {
			return errors.Join(errs.ErrMemo, err)
		}
	}
	return nil
}

func ValidateMemoDates(startDate, endDate time.Time) error {
	today := utils.NowPH().Format(constants.DateLayoutISO)
	startStr := startDate.Format(constants.DateLayoutISO)
	endStr := endDate.Format(constants.DateLayoutISO)
	if startStr < today || endStr < today {
		return errs.ErrMemoDateBeforeToday
	}
	if startDate.After(endDate) {
		return errs.ErrValidationStartEndDates
	}
	return nil
}

func (s *MemoService) decodeStaffIDs(encodedIDs []string) ([]int64, error) {
	result := make([]int64, 0, len(encodedIDs))
	seen := make(map[int64]struct{}, len(encodedIDs))
	for _, encodedID := range encodedIDs {
		id := s.encoder.Decode(encodedID)
		if id == encode.INVALID {
			return nil, errs.ErrDecode
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	if len(result) == 0 {
		return nil, errs.ErrMemoRecipientsRequired
	}
	return result, nil
}

func (s *MemoService) mapRowToMemo(m queries.TblMemo) Memo {
	createdAt, _ := time.Parse(constants.DateTimeLayoutISO, m.CreatedAt)
	var updatedAt sql.NullString
	if m.UpdatedAt != "" {
		updatedAt = sql.NullString{String: m.UpdatedAt, Valid: true}
	}
	return Memo{
		ID:           m.ID,
		Title:        m.Title,
		Message:      m.Message,
		FileURL:      m.FileUrl.String,
		Status:       enums.ParseMemoStatusToEnum(m.Status),
		StartDate:    m.StartDate,
		EndDate:      m.EndDate,
		CreatedBy:    m.CreatedBy,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    m.DeletedAt,
		EmailsSentAt: m.EmailsSentAt,
	}
}

func (s *MemoService) ID() string {
	return "Memo"
}

func (s *MemoService) Log() {
	logs.Log().Info("[MemoService] Loaded")
}

var _ IService = (*MemoService)(nil)
