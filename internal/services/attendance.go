package services

import (
	"context"
	"database/sql"
	"time"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/staff"
)

type AttendanceService struct {
	dbRO database.Service
	dbRW database.Service
}

func NewAttendanceService(ro, rw database.Service) *AttendanceService {
	return &AttendanceService{dbRO: ro, dbRW: rw}
}

func (s *AttendanceService) TimeIn(
	ctx context.Context,
	staffID int64,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx,
		queries.GetStaffAttendanceByDateParams{
			StaffID: staffID,
			ForDate: date,
		})

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == nil && existing.TimeIn.Valid {
		return sql.ErrNoRows
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx,
			queries.CreateStaffAttendanceParams{
				StaffID:       staffID,
				ForDate:       date,
				TimeIn:        sql.NullString{String: now, Valid: true},
				TimeOut:       sql.NullString{},
				InLocation:    location,
				InUseragentID: useragentID,
			})
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeIn(ctx,
		queries.UpdateStaffAttendanceTimeInParams{
			StaffID:       staffID,
			ForDate:       date,
			TimeIn:        sql.NullString{String: now, Valid: true},
			InLocation:    location,
			InUseragentID: useragentID,
		})

	return err
}

func (s *AttendanceService) TimeOut(
	ctx context.Context,
	staffID int64,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx,
		queries.GetStaffAttendanceByDateParams{
			StaffID: staffID,
			ForDate: date,
		})
	if err != nil {
		return err
	}

	if !existing.TimeIn.Valid {
		return sql.ErrNoRows
	}

	if existing.TimeOut.Valid {
		return sql.ErrTxDone
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeOut(ctx,
		queries.UpdateStaffAttendanceTimeOutParams{
			TimeOut:        sql.NullString{String: now, Valid: true},
			OutLocation:    location,
			OutUseragentID: useragentID,
			StaffID:        staffID,
			ForDate:        date,
		})

	return err
}

func (s *AttendanceService) LunchBreakIn(
	ctx context.Context,
	staffID int64,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	_, err := s.dbRW.GetQueries().UpdateStaffAttendanceLunchBreakIn(ctx, queries.UpdateStaffAttendanceLunchBreakInParams{
		LunchBreakIn:            sql.NullString{String: now, Valid: true},
		LunchBreakInLocation:    location,
		LunchBreakInUseragentID: useragentID,
		StaffID:                 staffID,
		ForDate:                 date,
	})
	return err
}

func (s *AttendanceService) LunchBreakOut(
	ctx context.Context,
	staffID int64,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
	if err != nil {
		return err
	}
	if !existing.LunchBreakIn.Valid {
		return sql.ErrNoRows
	}
	if existing.LunchBreakOut.Valid {
		return sql.ErrTxDone
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceLunchBreakOut(ctx, queries.UpdateStaffAttendanceLunchBreakOutParams{
		LunchBreakOut:            sql.NullString{String: now, Valid: true},
		LunchBreakOutLocation:    location,
		LunchBreakOutUseragentID: useragentID,
		StaffID:                  staffID,
		ForDate:                  date,
	})
	return err
}

func (s *AttendanceService) TimeOff(
	ctx context.Context,
	staffID int64,
	timeOffType enums.TimeOff,
	description string,
	startDate time.Time,
	endDate time.Time,
	useragentID sql.NullInt64,
) error {
	_, err := s.dbRW.GetQueries().CreateStaffTimeOff(
		ctx,
		queries.CreateStaffTimeOffParams{
			Type:        timeOffType.String(),
			StartDate:   startDate,
			EndDate:     endDate,
			Description: description,
			StaffID:     staffID,
			UseragentID: useragentID,
		},
	)
	return err
}

func (s *AttendanceService) UpsertLocation(
	ctx context.Context,
	staffID int64,
	date string,
	location sql.NullString,
) error {
	_, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx, queries.CreateStaffAttendanceParams{
			StaffID:     staffID,
			ForDate:     date,
			TimeIn:      sql.NullString{},
			TimeOut:     sql.NullString{},
			OutLocation: location,
		})
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceLocation(ctx, queries.UpdateStaffAttendanceLocationParams{
		OutLocation: location,
		StaffID:     staffID,
		ForDate:     date,
	})

	return err
}

func (s *AttendanceService) ApproveTimeOff(
	ctx context.Context,
	timeOffID int64,
	approvedBy int64,
) error {
	_, err := s.dbRW.GetQueries().ApproveStaffTimeOff(ctx, queries.ApproveStaffTimeOffParams{
		ApprovedBy: sql.NullInt64{Int64: approvedBy, Valid: true},
		ID:         timeOffID,
	})
	return err
}

func (s *AttendanceService) CancelTimeOff(
	ctx context.Context,
	timeOffID int64,
	approvedBy int64,
) error {
	_, err := s.dbRW.GetQueries().CancelStaffTimeOff(ctx, timeOffID)
	return err
}

func (s *AttendanceService) GetAttendance(
	ctx context.Context,
	staffID int64,
	startDate string,
	endDate string,
) ([]staff.StaffRow, error) {
	var data []staff.StaffRow
	if staffID != encode.INVALID {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRangeAndStaffID(ctx, queries.GetStaffAttendanceByDateRangeAndStaffIDParams{
			StaffID:   staffID,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]staff.StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, staff.StaffRow(att))
		}
	} else {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRange(ctx, queries.GetStaffAttendanceByDateRangeParams{
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]staff.StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, staff.StaffRow(att))
		}
	}
	return data, nil
}

