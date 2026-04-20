package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"github.com/gosimple/slug"
	"go.uber.org/zap"
)

type TrackLinkService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewTrackLinkService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *TrackLinkService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &TrackLinkService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *TrackLinkService) GetTrackedLinkByID(ctx context.Context, idStr string) (*TrackedLink, error) {
	link, err := s.dbRO.GetQueries().GetTrackedLinkByID(ctx, idStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return s.mapRowToTrackedLink(link), nil
}

func (s *TrackLinkService) GetTrackedLinkBySlug(ctx context.Context, slugs string) (*TrackedLink, error) {
	link, err := s.dbRO.GetQueries().GetTrackedLinkBySlug(ctx, slugs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s.mapRowToTrackedLink(link), nil
}

func (s *TrackLinkService) ListTrackedLinks(ctx context.Context) ([]TrackedLink, error) {
	links, err := s.dbRO.GetQueries().ListTrackedLinks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]TrackedLink, 0, len(links))
	for _, l := range links {
		result = append(result, *s.mapRowToTrackedLink(l))
	}
	return result, nil
}

func (s *TrackLinkService) CreateTrackedLink(
	ctx context.Context,
	staffID string,
	name string,
	slugs string,
	destinationURL string,
	source enums.TrackedLinkSource,
	medium enums.TrackedLinkMedium,
	campaign string,
) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModuleTrackedLinks,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	id := utils.GenString(16)
	staffIDNull := sql.NullString{String: staffID, Valid: staffID != ""}
	campaignNull := sql.NullString{String: campaign, Valid: campaign != ""}

	_, err := s.dbRW.GetQueries().CreateTrackedLink(ctx, queries.CreateTrackedLinkParams{
		ID:             id,
		StaffID:        staffIDNull,
		Name:           name,
		Slug:           slug.Make(slugs),
		DestinationUrl: destinationURL,
		Source:         sql.NullString{String: source.String(), Valid: source != enums.TRACKED_LINK_SOURCE_UNDEFINED},
		Medium:         sql.NullString{String: medium.String(), Valid: medium != enums.TRACKED_LINK_MEDIUM_UNDEFINED},
		Campaign:       campaignNull,
	})
	if err != nil {
		result = err.Error()
		return "", err
	}

	result = fmt.Sprintf("success. ID '%s'", id)
	return id, nil
}

func (s *TrackLinkService) UpdateTrackedLink(
	ctx context.Context,
	staffID string,
	id string,
	name string,
	slugs string,
	destinationURL string,
	source enums.TrackedLinkSource,
	medium enums.TrackedLinkMedium,
	campaign string,
	status enums.TrackedLinkStatus,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleTrackedLinks,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if err := s.dbRW.GetQueries().UpdateTrackedLink(ctx, queries.UpdateTrackedLinkParams{
		Name:           name,
		Slug:           slug.Make(slugs),
		DestinationUrl: destinationURL,
		Source:         sql.NullString{String: source.String(), Valid: source != enums.TRACKED_LINK_SOURCE_UNDEFINED},
		Medium:         sql.NullString{String: medium.String(), Valid: medium != enums.TRACKED_LINK_MEDIUM_UNDEFINED},
		Campaign:       sql.NullString{String: campaign, Valid: campaign != ""},
		Status:         status.String(),
		ID:             id,
	}); err != nil {
		result = err.Error()
		return err
	}

	result = fmt.Sprintf("success. ID '%s'", id)
	return nil
}

func (s *TrackLinkService) DeleteTrackedLink(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionDelete,
			constants.ModuleTrackedLinks,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()

	if err := s.dbRW.GetQueries().SoftDeleteTrackedLink(ctx, id); err != nil {
		result = err.Error()
		return err
	}

	result = fmt.Sprintf("success. ID '%s'", id)
	return nil
}

func (s *TrackLinkService) RecordClick(
	ctx context.Context,
	slugs string,
	referrer string,
	userAgent string,
	ipHash string,
	device string,
	utmSource string,
	utmMedium string,
	utmCampaign string,
) error {
	link, err := s.GetTrackedLinkBySlug(ctx, slugs)
	if err != nil {
		return err
	}
	if link == nil {
		return errs.ErrNotFound
	}

	err = s.dbRW.GetQueries().CreateLinkClick(ctx, queries.CreateLinkClickParams{
		LinkID:      link.ID,
		Referrer:    sql.NullString{String: referrer, Valid: referrer != ""},
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
		IpHash:      sql.NullString{String: ipHash, Valid: ipHash != ""},
		Device:      sql.NullString{String: device, Valid: device != ""},
		UtmSource:   sql.NullString{String: utmSource, Valid: utmSource != ""},
		UtmMedium:   sql.NullString{String: utmMedium, Valid: utmMedium != ""},
		UtmCampaign: sql.NullString{String: utmCampaign, Valid: utmCampaign != ""},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TrackLinkService) GetClickCount(ctx context.Context, linkID string) (int64, error) {
	count, err := s.dbRO.GetQueries().CountLinkClicksByLinkID(ctx, linkID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TrackLinkService) mapRowToTrackedLink(row queries.TblTrackedLink) *TrackedLink {
	return &TrackedLink{
		ID:             row.ID,
		Name:           row.Name,
		Slug:           row.Slug,
		DestinationURL: row.DestinationUrl,
		Source:         enums.ParseTrackedLinkSourceToEnum(row.Source.String),
		Medium:         enums.ParseTrackedLinkMediumToEnum(row.Medium.String),
		Campaign:       row.Campaign,
		Status:         enums.ParseTrackedLinkStatusToEnum(row.Status),
		StaffID:        row.StaffID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func (s *TrackLinkService) ID() string {
	return "TRACKED_LINK"
}

func (s *TrackLinkService) Log() {
	logs.Log().Info("[TrackLinkService] Loaded")
}

var _ IService = (*TrackLinkService)(nil)
