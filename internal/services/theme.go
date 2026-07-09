package services

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
)

type ThemeService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewThemeService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *ThemeService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ThemeService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *ThemeService) GetThemeByID(ctx context.Context, id int64) (*Theme, error) {
	t, err := s.dbRO.GetQueries().GetThemeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(errs.ErrTheme, err)
	}

	return s.mapRowToTheme(t.TblTheme), nil
}

func (s *ThemeService) GetAllThemes(ctx context.Context, search, sortBy, sortDir string) ([]Theme, error) {
	sortBy, sortDir = normalizeThemeListingSort(sortBy, sortDir)
	searchParam := sql.NullString{String: search, Valid: search != ""}

	themeRows, err := s.queryThemesForListing(ctx, sortBy, sortDir, searchParam)
	if err != nil {
		return nil, errors.Join(errs.ErrTheme, err)
	}

	result := make([]Theme, 0, len(themeRows))
	for _, row := range themeRows {
		theme := s.mapRowToTheme(row.tblTheme)
		theme.Active = row.active
		result = append(result, *theme)
	}
	return result, nil
}

func (s *ThemeService) queryThemesForListing(
	ctx context.Context,
	sortBy, sortDir string,
	search sql.NullString,
) ([]themeListingRow, error) {
	q := s.dbRO.GetQueries()

	switch sortBy {
	case "START_DATE":
		if sortDir == "ASC" {
			rows, err := q.SearchThemesSortStartDateAsc(ctx, search)
			if err != nil {
				return nil, err
			}
			return mapThemeRowsFromStartDateAsc(rows), nil
		}
		rows, err := q.SearchThemesSortStartDateDesc(ctx, search)
		if err != nil {
			return nil, err
		}
		return mapThemeRowsFromStartDateDesc(rows), nil
	case "END_DATE":
		if sortDir == "ASC" {
			rows, err := q.SearchThemesSortEndDateAsc(ctx, search)
			if err != nil {
				return nil, err
			}
			return mapThemeRowsFromEndDateAsc(rows), nil
		}
		rows, err := q.SearchThemesSortEndDateDesc(ctx, search)
		if err != nil {
			return nil, err
		}
		return mapThemeRowsFromEndDateDesc(rows), nil
	case "STATUS":
		if sortDir == "ASC" {
			rows, err := q.SearchThemesSortStatusAsc(ctx, search)
			if err != nil {
				return nil, err
			}
			return mapThemeRowsFromStatusAsc(rows), nil
		}
		rows, err := q.SearchThemesSortStatusDesc(ctx, search)
		if err != nil {
			return nil, err
		}
		return mapThemeRowsFromStatusDesc(rows), nil
	default:
		if sortDir == "DESC" {
			rows, err := q.SearchThemesSortTitleDesc(ctx, search)
			if err != nil {
				return nil, err
			}
			return mapThemeRowsFromTitleDesc(rows), nil
		}
		rows, err := q.SearchThemesSortTitleAsc(ctx, search)
		if err != nil {
			return nil, err
		}
		return mapThemeRowsFromTitleAsc(rows), nil
	}
}

func (s *ThemeService) GetActiveTheme(ctx context.Context) (*Theme, error) {
	t, err := s.dbRO.GetQueries().GetActiveTheme(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(errs.ErrTheme, err)
	}

	return s.mapRowToTheme(t.TblTheme), nil
}

func (s *ThemeService) CreateTheme(
	ctx context.Context,
	staffID string,
	title string,
	startDate time.Time,
	endDate time.Time,
	configuration map[string]string,
	configType enums.ThemeConfigType,
) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModuleThemes,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if !constants.ReThemeTitle.MatchString(title) {
		result = errs.ErrThemeInvalidTitle.Error()
		return "", errs.ErrThemeInvalidTitle
	}

	todayStr := time.Now().Format(constants.DateLayoutISO)
	if startDate.Format(constants.DateLayoutISO) < todayStr || endDate.Format(constants.DateLayoutISO) < todayStr {
		result = errs.ErrThemePastDate.Error()
		return "", errs.ErrThemePastDate
	}

	if startDate.After(endDate) {
		result = errs.ErrValidationStartEndDates.Error()
		return "", errs.ErrValidationStartEndDates
	}

	startDateStr := startDate.Format(constants.DateLayoutISO)
	endDateStr := endDate.Format(constants.DateLayoutISO)

	if err := s.assertNoOverlap(ctx, 0, startDateStr, endDateStr); err != nil {
		result = err.Error()
		return "", err
	}

	rawConfig, err := MarshalThemeConfiguration(configuration, configType)
	if err != nil {
		result = errs.ErrThemeInvalidConfig.Error()
		return "", errors.Join(errs.ErrThemeInvalidConfig, err)
	}

	createdBy := s.encoder.Decode(staffID)
	if createdBy == encode.INVALID {
		result = errs.ErrDecode.Error()
		return "", errs.ErrDecode
	}

	id, err := s.dbRW.GetQueries().CreateTheme(ctx, queries.CreateThemeParams{
		Title:             title,
		StartDate:         startDateStr,
		EndDate:           endDateStr,
		Configuration:     rawConfig,
		ConfigurationType: configType.String(),
		CreatedBy:         createdBy,
	})
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrTheme, err)
	}

	themeID := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", themeID)
	return themeID, nil
}

func (s *ThemeService) UpdateTheme(
	ctx context.Context,
	staffID string,
	themeID string,
	title string,
	status enums.ThemeStatus,
	startDate time.Time,
	endDate time.Time,
	configuration map[string]string,
	configType enums.ThemeConfigType,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleThemes,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := s.encoder.Decode(themeID)
	if id == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}
	existing, err := s.GetThemeByID(ctx, id)
	if err != nil {
		result = err.Error()
		return err
	}
	if existing == nil {
		result = errs.ErrThemeNotFound.Error()
		return errs.ErrThemeNotFound
	}

	if title != existing.Title && !constants.ReThemeTitle.MatchString(title) {
		result = errs.ErrThemeInvalidTitle.Error()
		return errs.ErrThemeInvalidTitle
	}

	startDateStr := startDate.Format(constants.DateLayoutISO)
	endDateStr := endDate.Format(constants.DateLayoutISO)
	datesChanged := startDateStr != existing.StartDate || endDateStr != existing.EndDate

	if datesChanged {
		todayStr := time.Now().Format(constants.DateLayoutISO)
		if startDateStr < todayStr || endDateStr < todayStr {
			result = errs.ErrThemePastDate.Error()
			return errs.ErrThemePastDate
		}
		if startDate.After(endDate) {
			result = errs.ErrValidationStartEndDates.Error()
			return errs.ErrValidationStartEndDates
		}
		if err := s.assertNoOverlap(ctx, id, startDateStr, endDateStr); err != nil {
			result = err.Error()
			return err
		}
	}

	rawConfig := existing.Configuration
	configChanged := configuration != nil && (configType != existing.ConfigurationType || !sameConfig(configuration, existing))
	if configChanged {
		marshaled, err := MarshalThemeConfiguration(configuration, configType)
		if err != nil {
			result = errs.ErrThemeInvalidConfig.Error()
			return errors.Join(errs.ErrThemeInvalidConfig, err)
		}
		rawConfig = marshaled
	}

	if err := s.dbRW.GetQueries().UpdateTheme(ctx, queries.UpdateThemeParams{
		ID:                id,
		Title:             cmp.Or(title, existing.Title),
		Status:            status.String(),
		StartDate:         startDateStr,
		EndDate:           endDateStr,
		Configuration:     rawConfig,
		ConfigurationType: configType.String(),
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrTheme, err)
	}

	result = fmt.Sprintf("success. ID '%s'", themeID)
	return nil
}

func (s *ThemeService) DeleteTheme(ctx context.Context, staffID string, themeID string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionDelete,
			constants.ModuleThemes,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := s.encoder.Decode(themeID)
	if id == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().SoftDeleteTheme(ctx, id); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrTheme, err)
	}

	result = fmt.Sprintf("success. ID '%s'", themeID)
	return nil
}

func (s *ThemeService) assertNoOverlap(ctx context.Context, excludeID int64, startDate, endDate string) error {
	overlapping, err := s.dbRO.GetQueries().GetOverlappingThemes(ctx, queries.GetOverlappingThemesParams{
		ID:        excludeID,
		StartDate: endDate,
		EndDate:   startDate,
	})
	if err != nil {
		return errors.Join(errs.ErrTheme, err)
	}
	if len(overlapping) > 0 {
		return errs.ErrThemeOverlappingDates
	}
	return nil
}

func (s *ThemeService) mapRowToTheme(t queries.TblTheme) *Theme {
	createdAt, _ := time.Parse(constants.DateTimeLayoutISO, t.CreatedAt)
	return &Theme{
		ID:                t.ID,
		Title:             t.Title,
		Status:            enums.ParseThemeStatusToEnum(t.Status),
		StartDate:         t.StartDate,
		EndDate:           t.EndDate,
		Configuration:     t.Configuration,
		ConfigurationType: enums.ParseThemeConfigTypeToEnum(t.ConfigurationType),
		CreatedBy:         t.CreatedBy,
		CreatedAt:         createdAt,
		UpdatedAt:         t.UpdatedAt,
		DeletedAt:         t.DeletedAt,
	}
}

func (s *ThemeService) ID() string {
	return "Theme"
}

func (s *ThemeService) Log() {
	logs.Log().Info("[ThemeService] Loaded")
}

var _ IService = (*ThemeService)(nil)
