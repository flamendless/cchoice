package services

import (
	"context"
	"database/sql"

	"cchoice/cmd/web/models"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"strings"

	"go.uber.org/zap"
)

type StaffLogsService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService
}

func NewStaffLogsService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
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
	logs.Log().Info(
		"StaffLogs",
		zap.String("staff id", staffID),
		zap.String("action", action),
		zap.String("module", module),
		zap.String("result", result),
		zap.Error(err),
	)
	return err
}

func (s *StaffLogsService) GetAll(ctx context.Context) ([]queries.GetAllStaffLogsRow, error) {
	return s.dbRO.GetQueries().GetAllStaffLogs(ctx)
}

func (s *StaffLogsService) GetDistinctActions(ctx context.Context) ([]string, error) {
	return s.dbRO.GetQueries().GetDistinctStaffLogActions(ctx)
}


func (s *StaffLogsService) GetFiltered(ctx context.Context, staffID int64, action string, module enums.Module) ([]queries.GetFilteredStaffLogsRow, error) {
	var moduleStr string
	if module.IsValid() {
		moduleStr = strings.ToLower(module.String())
	}
	return s.dbRO.GetQueries().GetFilteredStaffLogs(ctx, queries.GetFilteredStaffLogsParams{
		Action:  action,
		Module:  moduleStr,
		StaffID: staffID,
	})
}

func (s *StaffLogsService) GetByID(ctx context.Context, id int64) (queries.GetStaffLogByIDRow, error) {
	return s.dbRO.GetQueries().GetStaffLogByID(ctx, id)
}

func (s *StaffLogsService) GetAllAsModel(ctx context.Context) ([]models.StaffLog, error) {
	logsData, err := s.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	logsList := make([]models.StaffLog, 0, len(logsData))
	for _, l := range logsData {
		logsList = append(logsList, models.StaffLog{
			ID:         s.encoder.Encode(l.ID),
			StaffID:    s.encoder.Encode(l.StaffID),
			CreatedAt:  l.CreatedAt,
			Action:     l.Action,
			Module:     l.Module,
			Result:     l.Result,
			FirstName:  l.FirstName.String,
			MiddleName: l.MiddleName.String,
			LastName:   l.LastName.String,
		})
	}
	return logsList, nil
}

func (s *StaffLogsService) GetFilteredAsModel(ctx context.Context, staffID int64, action string, module enums.Module) ([]models.StaffLog, error) {
	logsData, err := s.GetFiltered(ctx, staffID, action, module)
	if err != nil {
		return nil, err
	}

	logsList := make([]models.StaffLog, 0, len(logsData))
	for _, l := range logsData {
		logsList = append(logsList, models.StaffLog{
			ID:         s.encoder.Encode(l.ID),
			StaffID:    s.encoder.Encode(l.StaffID),
			CreatedAt:  l.CreatedAt,
			Action:     l.Action,
			Module:     l.Module,
			Result:     l.Result,
			FirstName:  l.FirstName.String,
			MiddleName: l.MiddleName.String,
			LastName:   l.LastName.String,
		})
	}
	return logsList, nil
}

func (s *StaffLogsService) ID() string {
	return "StaffLogs"
}

func (s *StaffLogsService) Log() {
	logs.Log().Info("[StaffLogsService] Loaded")
}

var _ IService = (*StaffLogsService)(nil)
