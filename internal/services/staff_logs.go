package services

import (
	"context"
	"database/sql"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
)

type StaffLogsService struct {
	encoder encode.IEncode
	dbRO    database.Service
	dbRW    database.Service
}

func NewStaffLogsService(
	encoder encode.IEncode,
	dbRO database.Service,
	dbRW database.Service,
) *StaffLogsService {
	return &StaffLogsService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
	}
}

func (s *StaffLogsService) CreateLog(
	ctx context.Context,
	staffID string,
	action, module, result string,
	useragentID *int64,
) error {
	var useragentIDParam sql.NullInt64
	if useragentID != nil {
		useragentIDParam = sql.NullInt64{Int64: *useragentID, Valid: true}
	}
	_, err := s.dbRW.GetQueries().CreateStaffLog(ctx, queries.CreateStaffLogParams{
		StaffID:     s.encoder.Decode(staffID),
		Action:      action,
		Module:      module,
		Result:      result,
		UseragentID: useragentIDParam,
	})
	return err
}

func (s *StaffLogsService) GetAll(ctx context.Context) ([]queries.GetAllStaffLogsRow, error) {
	return s.dbRO.GetQueries().GetAllStaffLogs(ctx)
}

func (s *StaffLogsService) GetByID(ctx context.Context, id int64) (queries.GetStaffLogByIDRow, error) {
	return s.dbRO.GetQueries().GetStaffLogByID(ctx, id)
}
