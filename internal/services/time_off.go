package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func sumHoursFromSQLCValue(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int64:
		return float64(n), nil
	case int:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("unexpected sum type %T", v)
	}
}

func (s *AttendanceService) LeaveCreditBalance(ctx context.Context, staffID string) (float64, error) {
	dbStaffID := s.encoder.Decode(staffID)
	grantedRaw, err := s.dbRO.GetQueries().SumStaffLeaveCreditHours(ctx, dbStaffID)
	if err != nil {
		return 0, err
	}
	usedRaw, err := s.dbRO.GetQueries().SumApprovedStaffTimeOffHours(ctx, dbStaffID)
	if err != nil {
		return 0, err
	}
	granted, err := sumHoursFromSQLCValue(grantedRaw)
	if err != nil {
		return 0, err
	}
	used, err := sumHoursFromSQLCValue(usedRaw)
	if err != nil {
		return 0, err
	}
	return granted - used, nil
}

func (s *AttendanceService) GrantLeaveCredits(
	ctx context.Context,
	staffID string,
	grantedByStaffID string,
	hours float64,
	note string,
) error {
	if hours <= 0 {
		return errs.ErrInvalidLeaveHours
	}
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			grantedByStaffID,
			constants.ActionGrant,
			constants.ModuleTimeOff,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	var noteVal sql.NullString
	if note != "" {
		noteVal = sql.NullString{String: note, Valid: true}
	}
	_, err := s.dbRW.GetQueries().CreateStaffLeaveCreditGrant(ctx, queries.CreateStaffLeaveCreditGrantParams{
		StaffID:   s.encoder.Decode(staffID),
		Hours:     hours,
		Note:      noteVal,
		GrantedBy: s.encoder.Decode(grantedByStaffID),
	})
	if err != nil {
		result = err.Error()
	}
	return err
}

func mapStaffTimeOffStatus(approved sql.NullBool, status string) enums.TimeOffStatus {
	parsed := enums.ParseTimeOffStatusToEnum(status)
	if parsed != enums.TIME_OFF_STATUS_UNDEFINED {
		return parsed
	}
	if approved.Valid && approved.Bool {
		return enums.TIME_OFF_STATUS_APPROVED
	}
	return enums.TIME_OFF_STATUS_PENDING
}

func buildStaffTimeOffModel(
	encoder encode.IEncode,
	id int64,
	staffDBID int64,
	fullName string,
	typeStr string,
	startDate time.Time,
	endDate time.Time,
	hours float64,
	description string,
	status string,
	approved sql.NullBool,
	rejectReason sql.NullString,
	approverFirst sql.NullString,
	approverMiddle sql.NullString,
	approverLast sql.NullString,
	approvedAt sql.NullTime,
	createdAt string,
) models.StaffTimeOff {
	var approvedBy string
	if approverFirst.Valid {
		approvedBy = utils.BuildFullName(approverFirst.String, approverMiddle.String, approverLast.String)
	} else {
		approvedBy = "-"
	}
	var approvedAtStr string
	if approvedAt.Valid {
		approvedAtStr = utils.FormatTimePH(approvedAt.Time)
	} else {
		approvedAtStr = "-"
	}
	rejectReasonStr := "-"
	if rejectReason.Valid && rejectReason.String != "" {
		rejectReasonStr = rejectReason.String
	}
	st := mapStaffTimeOffStatus(approved, status)
	staffIDEncoded := ""
	if staffDBID != 0 {
		staffIDEncoded = encoder.Encode(staffDBID)
	}
	return models.StaffTimeOff{
		ID:           encoder.Encode(id),
		StaffID:      staffIDEncoded,
		FullName:     fullName,
		Type:         enums.ParseTimeOffToEnum(typeStr),
		CreatedAt:    utils.ConvertToPH(createdAt),
		StartDate:    startDate.Format(constants.DateLayoutISO),
		EndDate:      endDate.Format(constants.DateLayoutISO),
		Hours:        hours,
		Description:  description,
		Status:       st,
		RejectReason: rejectReasonStr,
		Approved:     st == enums.TIME_OFF_STATUS_APPROVED,
		ApprovedBy:   approvedBy,
		ApprovedAt:   approvedAtStr,
	}
}
