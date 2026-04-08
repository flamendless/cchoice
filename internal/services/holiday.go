package services

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type Holiday struct {
	ID        int64
	Date      string            `json:"date"`
	Name      string            `json:"name"`
	Type      enums.HolidayType `json:"type"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt sql.NullString    `json:"updated_at"`
}

type HolidayService struct {
	encoder encode.IEncode
	dbRO    database.IService
	dbRW    database.IService

	mu    sync.RWMutex
	cache map[string]*Holiday
}

func NewHolidayService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
) *HolidayService {
	s := &HolidayService{
		encoder: encoder,
		dbRO:    dbRO,
		dbRW:    dbRW,
		cache:   make(map[string]*Holiday),
	}
	s.loadCache(context.Background())
	return s
}

func (s *HolidayService) loadCache(ctx context.Context) {
	holidays, err := s.dbRO.GetQueries().GetAllHolidays(ctx)
	if err != nil {
		logs.Log().Error("failed to load holidays cache", zap.Error(err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*Holiday, len(holidays))
	for _, h := range holidays {
		ht := enums.ParseHolidayTypeToEnum(h.Type)
		createdAt, _ := time.Parse(constants.DateTimeLayoutISO, h.CreatedAt)
		s.cache[h.Date] = &Holiday{
			Date:      h.Date,
			Name:      h.Name,
			Type:      ht,
			CreatedAt: createdAt,
			UpdatedAt: h.UpdatedAt,
		}
	}
	logs.Log().Info("holidays cache loaded", zap.Int("count", len(holidays)))
}

func (s *HolidayService) RefreshCache(ctx context.Context) {
	s.loadCache(ctx)
}

func (s *HolidayService) IsHoliday(ctx context.Context, date time.Time) (*Holiday, error) {
	dateStr := date.Format(constants.DateLayoutISO)

	s.mu.RLock()
	holiday, exists := s.cache[dateStr]
	s.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	return &Holiday{
		Date:      holiday.Date,
		Name:      holiday.Name,
		Type:      holiday.Type,
		CreatedAt: holiday.CreatedAt,
		UpdatedAt: holiday.UpdatedAt,
	}, nil
}

func (s *HolidayService) GetAllHolidays(ctx context.Context) ([]Holiday, error) {
	holidays, err := s.dbRO.GetQueries().GetAllHolidays(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrHoliday, err)
	}

	result := make([]Holiday, 0, len(holidays))
	for _, h := range holidays {
		ht := enums.ParseHolidayTypeToEnum(h.Type)
		createdAt, _ := time.Parse(constants.DateTimeLayoutISO, h.CreatedAt)
		result = append(result, Holiday{
			ID:        h.ID,
			Date:      h.Date,
			Name:      h.Name,
			Type:      ht,
			CreatedAt: createdAt,
			UpdatedAt: h.UpdatedAt,
		})
	}
	return result, nil
}

func (s *HolidayService) GetHolidaysByDateRange(ctx context.Context, startDate, endDate time.Time) ([]Holiday, error) {
	holidays, err := s.dbRO.GetQueries().GetHolidaysByDateRange(ctx, queries.GetHolidaysByDateRangeParams{
		StartDate: startDate.Format(constants.DateLayoutISO),
		EndDate:   endDate.Format(constants.DateLayoutISO),
	})
	if err != nil {
		return nil, errors.Join(errs.ErrHoliday, err)
	}

	result := make([]Holiday, 0, len(holidays))
	for _, h := range holidays {
		ht := enums.ParseHolidayTypeToEnum(h.Type)
		createdAt, _ := time.Parse(constants.DateTimeLayoutISO, h.CreatedAt)
		result = append(result, Holiday{
			Date:      h.Date,
			Name:      h.Name,
			Type:      ht,
			CreatedAt: createdAt,
			UpdatedAt: h.UpdatedAt,
		})
	}
	return result, nil
}

func (s *HolidayService) CreateHoliday(ctx context.Context, date time.Time, name string, holidayType enums.HolidayType) (int64, error) {
	id, err := s.dbRW.GetQueries().CreateHoliday(ctx, queries.CreateHolidayParams{
		Date: date.Format(constants.DateLayoutISO),
		Name: name,
		Type: holidayType.String(),
	})
	if err != nil {
		return 0, errors.Join(errs.ErrHoliday, err)
	}
	s.RefreshCache(ctx)
	return id, nil
}

func (s *HolidayService) UpdateHoliday(ctx context.Context, id int64, name string, holidayType enums.HolidayType) (int64, error) {
	id, err := s.dbRW.GetQueries().UpdateHoliday(ctx, queries.UpdateHolidayParams{
		Name: name,
		Type: holidayType.String(),
		ID:   id,
	})
	if err != nil {
		return 0, errors.Join(errs.ErrHoliday, err)
	}
	s.RefreshCache(ctx)
	return id, nil
}

func (s *HolidayService) DeleteHoliday(ctx context.Context, id int64) error {
	err := s.dbRW.GetQueries().DeleteHoliday(ctx, id)
	if err != nil {
		return errors.Join(errs.ErrHoliday, err)
	}
	s.RefreshCache(ctx)
	return nil
}

func (s *HolidayService) Log() {
	logs.Log().Info("[HolidayService] Loaded")
}

var _ IService = (*HolidayService)(nil)
