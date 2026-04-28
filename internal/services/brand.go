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

	"go.uber.org/zap"
)

type BrandService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewBrandService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *BrandService {
	return &BrandService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *BrandService) GetNameByID(ctx context.Context, brandID string) (string, error) {
	brand, err := s.dbRO.GetQueries().GetBrandsByID(ctx, s.encoder.Decode(brandID))
	if err != nil {
		return "", err
	}
	return brand.Name, nil
}

func (s *BrandService) GetAllBrands(ctx context.Context) ([]Brand, error) {
	brands, err := s.dbRO.GetQueries().GetAllBrands(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	result := make([]Brand, 0, len(brands))
	for _, b := range brands {
		s3URL := ""
		if b.S3Url.Valid {
			s3URL = b.S3Url.String
		}
		brandImageID := int64(0)
		if b.BrandImageID.Valid {
			brandImageID = b.BrandImageID.Int64
		}
		result = append(result, Brand{
			ID:           b.ID,
			Name:         b.Name,
			LogoS3URL:    s3URL,
			BrandImageID: brandImageID,
			ProductCount: b.ProductCount,
			CreatedAt:    b.CreatedAt,
			Status:       enums.ParseBrandStatusToEnum(b.Status),
		})
	}
	return result, nil
}

func (s *BrandService) SearchBrandsByFilter(
	ctx context.Context,
	name string,
	status enums.BrandStatus,
) ([]Brand, error) {
	var statusStr string
	if status != enums.BRAND_STATUS_UNDEFINED {
		statusStr = status.String()
	}
	brands, err := s.dbRO.GetQueries().SearchBrandsByFilter(ctx, queries.SearchBrandsByFilterParams{
		SearchName: name,
		Status:     statusStr,
	})
	if err != nil {
		return nil, errors.Join(errs.ErrBrand, err)
	}

	result := make([]Brand, 0, len(brands))
	for _, b := range brands {
		s3URL := ""
		if b.S3Url.Valid {
			s3URL = b.S3Url.String
		}
		brandImageID := int64(0)
		if b.BrandImageID.Valid {
			brandImageID = b.BrandImageID.Int64
		}
		result = append(result, Brand{
			ID:           b.ID,
			Name:         b.Name,
			LogoS3URL:    s3URL,
			BrandImageID: brandImageID,
			ProductCount: b.ProductCount,
			CreatedAt:    b.CreatedAt,
		})
	}
	return result, nil
}

func (s *BrandService) CreateBrand(ctx context.Context, staffID string, name string, logoS3URL string) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionCreate,
			constants.ModuleBrands,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[BrandService] create log", zap.Error(err))
		}
	}()

	brandID, err := s.dbRW.GetQueries().CreateBrands(ctx, name)
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrBrand, err)
	}

	brandIDStr := s.encoder.Encode(brandID)
	if _, err = s.dbRW.GetQueries().CreateBrandImages(ctx, queries.CreateBrandImagesParams{
		BrandID: brandID,
		Path:    "",
		S3Url:   sql.NullString{String: logoS3URL, Valid: logoS3URL != ""},
		IsMain:  true,
	}); err != nil {
		result = err.Error()
		return brandIDStr, errors.Join(errs.ErrBrand, err)
	}

	result = fmt.Sprintf("success. ID '%s'", brandIDStr)
	return brandIDStr, nil
}

func (s *BrandService) UpdateBrand(ctx context.Context, staffID string, id string, name string, logoS3URL string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleBrands,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[BrandService] update log", zap.Error(err))
		}
	}()

	brandID := s.encoder.Decode(id)
	if err := s.dbRW.GetQueries().UpdateBrand(ctx, queries.UpdateBrandParams{
		ID:   brandID,
		Name: name,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrBrand, err)
	}

	if logoS3URL != "" {
		if err := s.dbRW.GetQueries().UpdateBrandImage(ctx, queries.UpdateBrandImageParams{
			BrandID: brandID,
			S3Url:   sql.NullString{String: logoS3URL, Valid: true},
		}); err != nil {
			result = err.Error()
			return errors.Join(errs.ErrBrand, err)
		}
	}

	return nil
}

func (s *BrandService) DeleteBrand(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionDelete,
			constants.ModuleBrands,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[BrandService] delete log", zap.Error(err))
		}
	}()

	brandID := s.encoder.Decode(id)
	if brandID == encode.INVALID {
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().SoftDeleteBrand(ctx, brandID); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrBrand, err)
	}

	return nil
}

func (s *BrandService) UpdateStatus(ctx context.Context, staffID string, brandID string, status enums.BrandStatus) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleBrands,
			result,
			nil,
		); err != nil {
			logs.Log().Warn("[BrandService] update status log", zap.Error(err))
		}
	}()

	decodedID := s.encoder.Decode(brandID)
	if decodedID == encode.INVALID {
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().UpdateBrandStatus(ctx, queries.UpdateBrandStatusParams{
		ID:     decodedID,
		Status: status.String(),
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrBrand, err)
	}

	result = fmt.Sprintf("success. ID '%s'", brandID)
	return nil
}

func (s *BrandService) ID() string {
	return "Brand"
}

func (s *BrandService) Log() {
	logs.Log().Info("[BrandService] Loaded")
}

var _ IService = (*BrandService)(nil)
