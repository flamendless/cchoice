package services

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
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

type PromoService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewPromoService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *PromoService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &PromoService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *PromoService) GetPromoByID(ctx context.Context, id int64) (*Promo, error) {
	p, err := s.dbRO.GetQueries().GetPromoByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Join(errs.ErrPromo, err)
	}

	return s.mapRowToPromo(p.TblPromo), nil
}

func (s *PromoService) GetAllPromos(ctx context.Context) ([]Promo, error) {
	promos, err := s.dbRO.GetQueries().GetAllPromos(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrPromo, err)
	}

	result := make([]Promo, 0, len(promos))
	for _, p := range promos {
		result = append(result, *s.mapRowToPromo(p.TblPromo))
	}
	return result, nil
}

func (s *PromoService) GetActivePromos(ctx context.Context) ([]Promo, error) {
	promos, err := s.dbRO.GetQueries().GetActivePromos(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrPromo, err)
	}

	result := make([]Promo, 0, len(promos))
	for _, p := range promos {
		result = append(result, *s.mapRowToPromo(p.TblPromo))
	}
	return result, nil
}

func (s *PromoService) CreatePromo(
	ctx context.Context,
	staffID string,
	title string,
	description string,
	mediaURL string,
	startDate time.Time,
	endDate time.Time,
	promoType enums.PromoType,
	bannerOnly bool,
	priority int64,
) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModulePromos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if startDate.After(endDate) {
		result = errs.ErrValidationStartEndDates.Error()
		return "", errs.ErrValidationStartEndDates
	}

	id, err := s.dbRW.GetQueries().CreatePromo(ctx, queries.CreatePromoParams{
		Title:       title,
		Description: description,
		MediaUrl:    mediaURL,
		StartDate:   startDate.Format(constants.DateLayoutISO),
		EndDate:     endDate.Format(constants.DateLayoutISO),
		Type:        promoType.String(),
		BannerOnly:  sql.NullBool{Valid: true, Bool: bannerOnly},
		Priority:    sql.NullInt64{Valid: true, Int64: priority},
	})
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrPromo, err)
	}

	promoID := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", promoID)
	return promoID, nil
}

func (s *PromoService) UpdatePromo(
	ctx context.Context,
	staffID string,
	promoID string,
	title string,
	description string,
	mediaURL string,
	startDate time.Time,
	endDate time.Time,
	promoType enums.PromoType,
	promoStatus enums.PromoStatus,
	bannerOnly bool,
	priority int64,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModulePromos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := s.encoder.Decode(promoID)
	promo, err := s.GetPromoByID(ctx, id)
	if err != nil {
		result = err.Error()
		return err
	}
	if promo == nil {
		result = errs.ErrPromo.Error()
		return errs.ErrPromo
	}

	if startDate.After(endDate) {
		result = errs.ErrValidationStartEndDates.Error()
		return errs.ErrValidationStartEndDates
	}

	if err := s.dbRW.GetQueries().UpdatePromo(ctx, queries.UpdatePromoParams{
		ID:          id,
		Title:       title,
		Description: description,
		MediaUrl:    cmp.Or(mediaURL, promo.MediaURL),
		StartDate:   startDate.Format(constants.DateLayoutISO),
		EndDate:     endDate.Format(constants.DateLayoutISO),
		Type:        promoType.String(),
		Status:      promoStatus.String(),
		BannerOnly:  sql.NullBool{Valid: true, Bool: bannerOnly},
		Priority:    sql.NullInt64{Valid: true, Int64: priority},
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrPromo, err)
	}

	result = fmt.Sprintf("success. ID '%s'", promoID)
	return nil
}

func (s *PromoService) DeletePromo(ctx context.Context, staffID string, promoID string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionDelete,
			constants.ModulePromos,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := s.encoder.Decode(promoID)
	if id == encode.INVALID {
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().SoftDeletePromo(ctx, id); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrPromo, err)
	}

	result = fmt.Sprintf("success. ID '%s'", promoID)
	return nil
}

func (s *PromoService) mapRowToPromo(p queries.TblPromo) *Promo {
	createdAt, _ := time.Parse(constants.DateTimeLayoutISO, p.CreatedAt)
	var updatedAt sql.NullString
	if p.UpdatedAt != "" {
		updatedAt = sql.NullString{String: p.UpdatedAt, Valid: true}
	}
	return &Promo{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		MediaURL:    p.MediaUrl,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		Type:        enums.ParsePromoTypeToEnum(p.Type),
		Status:      enums.ParsePromoStatusToEnum(p.Status),
		BannerOnly:  p.BannerOnly,
		Priority:    p.Priority,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   p.DeletedAt,
	}
}

func (s *PromoService) ID() string {
	return "Promo"
}

func (s *PromoService) Log() {
	logs.Log().Info("[PromoService] Loaded")
}

var _ IService = (*PromoService)(nil)
