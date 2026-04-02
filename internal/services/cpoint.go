package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type CpointService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewCpointService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *CpointService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &CpointService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

type CreateCpointParams struct {
	StaffID     string
	CustomerID  string
	Value       int64
	ProductSkus []string
	ExpiresAt   *time.Time
}

type Cpoint struct {
	ID          int64
	CustomerID  int64
	Code        string
	Value       int64
	ProductSkus []string
	ExpiresAt   *time.Time
	GeneratedAt time.Time
	RedeemedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetCpointsByCustomerIDRowWithTotal struct {
	Cpoint
	Total int64
}

const prefix = "CP"
const cpointCodeChars = "ABCDEFGHJKMNPQRSTUVWXYZ123456789"

func (s *CpointService) GenerateCode() string {
	segments := []string{prefix, "", "", ""}
	for i := 1; i <= 3; i++ {
		segment := make([]byte, 3)
		for j := range segment {
			b := make([]byte, 1)
			if _, err := rand.Read(b); err != nil {
				logs.Log().Info("rand", zap.Error(err))
			}
			num := int(b[0]) % len(cpointCodeChars)
			segment[j] = cpointCodeChars[num]
		}
		segments[i] = string(segment)
	}
	return strings.Join(segments, "-")
}

func (s *CpointService) CreateCpoint(ctx context.Context, params CreateCpointParams) (int64, error) {
	var result string
	defer func() {
		if err := s.staffLog.CreateLog(ctx, params.StaffID, "CREATE_CPOINT", "CPOINTS", result, nil); err != nil {
			logs.Log().Warn("create log", zap.Error(err))
		}
	}()
	customerIDDecoded := s.encoder.Decode(params.CustomerID)
	if customerIDDecoded == encode.INVALID {
		return 0, errs.ErrDecode
	}

	code := s.GenerateCode()

	var productSkusStr sql.NullString
	if len(params.ProductSkus) > 0 {
		productSkusStr = sql.NullString{String: strings.Join(params.ProductSkus, ","), Valid: true}
	}

	var expiresAtStr sql.NullString
	if params.ExpiresAt != nil {
		expiresAtStr = sql.NullString{String: params.ExpiresAt.Format(time.RFC3339), Valid: true}
	} else {
		oneYearLater := time.Now().AddDate(1, 0, 0)
		expiresAtStr = sql.NullString{String: oneYearLater.Format(time.RFC3339), Valid: true}
	}

	cpointID, err := s.dbRW.GetQueries().CreateCpoint(ctx, queries.CreateCpointParams{
		CustomerID:  customerIDDecoded,
		Code:        code,
		Value:       params.Value,
		ProductSkus: productSkusStr,
		ExpiresAt:   expiresAtStr,
	})
	if err != nil {
		result = fmt.Sprintf("FAILURE: %v", err)
		return 0, err
	}

	result = fmt.Sprintf("SUCCESS: %d", cpointID)

	return cpointID, nil
}

func (s *CpointService) RedeemCpoint(ctx context.Context, code string) error {
	cpoint, err := s.GetCpointByCode(ctx, code)
	if err != nil {
		return err
	}
	if cpoint.RedeemedAt != nil {
		return errs.ErrCpointAlreadyRedeemed
	}
	if cpoint.ExpiresAt != nil && cpoint.ExpiresAt.Before(time.Now()) {
		return errs.ErrCpointExpired
	}

	if _, err = s.dbRW.GetQueries().RedeemCpoint(ctx, code); err != nil {
		return err
	}

	return nil
}

func (s *CpointService) GetCpointByCode(ctx context.Context, code string) (Cpoint, error) {
	row, err := s.dbRO.GetQueries().GetCpointByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Cpoint{}, errs.ErrCpointNotFound
		}
		return Cpoint{}, err
	}

	var expiresAt *time.Time
	if row.ExpiresAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.ExpiresAt.String)
		expiresAt = &t
	}

	var redeemedAt *time.Time
	if row.RedeemedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.RedeemedAt.String)
		redeemedAt = &t
	}

	generatedAt, _ := time.Parse(time.RFC3339, row.GeneratedAt)
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)

	productSkus := []string{}
	if row.ProductSkus.Valid {
		productSkus = strings.Split(row.ProductSkus.String, ",")
	}

	return Cpoint{
		ID:          row.ID,
		CustomerID:  row.CustomerID,
		Code:        row.Code,
		Value:       row.Value,
		ProductSkus: productSkus,
		ExpiresAt:   expiresAt,
		GeneratedAt: generatedAt,
		RedeemedAt:  redeemedAt,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func (s *CpointService) GetCpointsByCustomerID(ctx context.Context, customerID string, withTotal bool) ([]GetCpointsByCustomerIDRowWithTotal, error) {
	customerIDDecoded := s.encoder.Decode(customerID)

	var rows []queries.GetCpointsByCustomerIDWithTotalRow
	var err error

	if withTotal {
		rows, err = s.dbRO.GetQueries().GetCpointsByCustomerIDWithTotal(ctx, customerIDDecoded)
	} else {
		rawRows, err := s.dbRO.GetQueries().GetCpointsByCustomerID(ctx, customerIDDecoded)
		if err != nil {
			return nil, err
		}
		for _, r := range rawRows {
			rows = append(rows, queries.GetCpointsByCustomerIDWithTotalRow{
				ID:          r.ID,
				CustomerID:  r.CustomerID,
				Code:        r.Code,
				Value:       r.Value,
				ProductSkus: r.ProductSkus,
				ExpiresAt:   r.ExpiresAt,
				GeneratedAt: r.GeneratedAt,
				RedeemedAt:  r.RedeemedAt,
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
				Total:       0,
			})
		}
	}
	if err != nil {
		return nil, err
	}

	result := make([]GetCpointsByCustomerIDRowWithTotal, 0, len(rows))
	for _, r := range rows {
		var expiresAt *time.Time
		if r.ExpiresAt.Valid {
			t, _ := time.Parse(time.RFC3339, r.ExpiresAt.String)
			expiresAt = &t
		}

		var redeemedAt *time.Time
		if r.RedeemedAt.Valid {
			t, _ := time.Parse(time.RFC3339, r.RedeemedAt.String)
			redeemedAt = &t
		}

		generatedAt, _ := time.Parse(time.RFC3339, r.GeneratedAt)
		createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)

		productSkus := []string{}
		if r.ProductSkus.Valid {
			productSkus = strings.Split(r.ProductSkus.String, ",")
		}

		result = append(result, GetCpointsByCustomerIDRowWithTotal{
			Cpoint: Cpoint{
				ID:          r.ID,
				CustomerID:  r.CustomerID,
				Code:        r.Code,
				Value:       r.Value,
				ProductSkus: productSkus,
				ExpiresAt:   expiresAt,
				GeneratedAt: generatedAt,
				RedeemedAt:  redeemedAt,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			Total: r.Total,
		})
	}

	return result, nil
}

func (s *CpointService) GenerateRedemptionURL(code string) string {
	return utils.URL("/cpoints/redeem?code=" + code)
}

func (s *CpointService) ValidateCode(code string) error {
	if len(code) != 14 {
		return errs.ErrInvalidInput
	}

	parts := strings.Split(code, "-")
	if len(parts) != 4 {
		return errs.ErrInvalidInput
	}

	if parts[0] != prefix {
		return errs.ErrInvalidInput
	}

	validChars := cpointCodeChars
	for _, part := range parts[1:] {
		for _, c := range part {
			if !strings.Contains(validChars, string(c)) {
				return errs.ErrInvalidInput
			}
		}
	}

	return nil
}
