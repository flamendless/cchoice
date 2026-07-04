package services

import (
	"context"
	"database/sql"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
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
		staffLog := s.toStaffLogModel(ctx, l.ID, l.StaffID, l.CreatedAt, l.Action, l.Module, l.Result, l.FirstName, l.MiddleName, l.LastName)
		logsList = append(logsList, staffLog)
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
		staffLog := s.toStaffLogModel(ctx, l.ID, l.StaffID, l.CreatedAt, l.Action, l.Module, l.Result, l.FirstName, l.MiddleName, l.LastName)
		logsList = append(logsList, staffLog)
	}
	return logsList, nil
}

func (s *StaffLogsService) GetFilteredAsModelPaginated(
	ctx context.Context,
	staffID int64,
	action string,
	module enums.Module,
	page, perPage int,
) ([]models.StaffLog, int64, int, error) {
	var moduleStr string
	if module.IsValid() {
		moduleStr = strings.ToLower(module.String())
	}

	filterParams := queries.CountFilteredStaffLogsParams{
		Action:  action,
		Module:  moduleStr,
		StaffID: staffID,
	}

	totalCount, err := s.dbRO.GetQueries().CountFilteredStaffLogs(ctx, filterParams)
	if err != nil {
		return nil, 0, 0, err
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	logsData, err := s.dbRO.GetQueries().GetFilteredStaffLogsPaginated(ctx, queries.GetFilteredStaffLogsPaginatedParams{
		Action:  action,
		Module:  moduleStr,
		StaffID: staffID,
		Limit:   int64(perPage),
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, 0, err
	}

	logsList := make([]models.StaffLog, 0, len(logsData))
	for _, l := range logsData {
		staffLog := s.toStaffLogModel(ctx, l.ID, l.StaffID, l.CreatedAt, l.Action, l.Module, l.Result, l.FirstName, l.MiddleName, l.LastName)
		logsList = append(logsList, staffLog)
	}
	return logsList, totalCount, page, nil
}

func (s *StaffLogsService) toStaffLogModel(
	ctx context.Context,
	id, staffID int64,
	createdAt, action, module, result string,
	firstName, middleName, lastName sql.NullString,
) models.StaffLog {
	staffLog := models.StaffLog{
		ID:         s.encoder.Encode(id),
		StaffID:    s.encoder.Encode(staffID),
		CreatedAt:  createdAt,
		Action:     action,
		Module:     module,
		Result:     result,
		FirstName:  firstName.String,
		MiddleName: middleName.String,
		LastName:   lastName.String,
	}
	s.enrichReference(ctx, &staffLog)
	return staffLog
}

func (s *StaffLogsService) enrichReference(ctx context.Context, log *models.StaffLog) {
	if log.Module != constants.ModuleProducts {
		return
	}
	if log.Action != constants.ActionCreate && log.Action != constants.ActionUpdate {
		return
	}
	encodedID, ok := utils.ParseStaffLogSuccessID(log.Result)
	if !ok {
		return
	}
	decoded := s.encoder.Decode(encodedID)
	if decoded == encode.INVALID {
		return
	}
	product, err := s.dbRO.GetQueries().GetProductSlugByID(ctx, decoded)
	if err != nil {
		return
	}
	slug := ""
	if product.Slug.Valid {
		slug = product.Slug.String
	}
	ref := utils.BuildStaffLogProductReference(slug, product.Serial, product.Status)
	if ref.Label == "" {
		return
	}
	log.ReferenceLabel = ref.Label
	log.ReferenceURL = ref.URL
	log.ReferenceNewTab = ref.NewTab
}

func (s *StaffLogsService) ID() string {
	return "StaffLogs"
}

func (s *StaffLogsService) Log() {
	logs.Log().Info("[StaffLogsService] Loaded")
}

var _ IService = (*StaffLogsService)(nil)
