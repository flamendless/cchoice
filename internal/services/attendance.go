package services

import (
	"context"
	"database/sql"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
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
			TimeIn:        sql.NullString{String: now, Valid: true},
			InLocation:    location,
			InUseragentID: useragentID,
			StaffID:       staffID,
			ForDate:       date,
		})

	return err
}

func (s *AttendanceService) TimeOut(
	ctx context.Context,
	staffID int64,
	date string,
	now string,
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
			OutUseragentID: useragentID,
			StaffID:        staffID,
			ForDate:        date,
		})

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
