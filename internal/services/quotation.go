package services

import (
	"context"
	"database/sql"
	"errors"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type QuotationService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	dbRW     database.IService
	staffLog *StaffLogsService
}

func NewQuotationService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
) *QuotationService {
	return &QuotationService{
		encoder:  encoder,
		dbRO:     dbRO,
		dbRW:     dbRW,
		staffLog: staffLog,
	}
}

func (s *QuotationService) GetOrCreateActive(ctx context.Context, customerID string) (queries.TblQuotation, error) {
	const logtag = "[QuotationService] GetOrCreateActive"
	decodedID := s.encoder.Decode(customerID)
	if decodedID == encode.INVALID {
		return queries.TblQuotation{}, errs.ErrDecode
	}

	quotation, err := s.dbRO.GetQueries().GetActiveQuotationByCustomerID(ctx, decodedID)
	if err == sql.ErrNoRows {
		quotation, err = s.dbRW.GetQueries().CreateQuotation(ctx, queries.CreateQuotationParams{
			CustomerID:            decodedID,
			AcknowledgedByStaffID: sql.NullInt64{Valid: false},
		})
		if err != nil {
			return queries.TblQuotation{}, err
		}
		logs.LogCtx(ctx).Info(logtag, zap.String("customer_id", customerID))
		return quotation, nil
	}
	if err != nil {
		return queries.TblQuotation{}, err
	}
	return quotation, nil
}

func (s *QuotationService) AddLineToQuotation(ctx context.Context, customerID string, productID string, quantity int64) error {
	const logtag = "[QuotationService] AddLineToQuotation"
	decodedProductID := s.encoder.Decode(productID)
	if decodedProductID == encode.INVALID {
		return errs.ErrDecode
	}

	quotation, err := s.GetOrCreateActive(ctx, customerID)
	if err != nil {
		return err
	}

	product, err := s.dbRO.GetQueries().GetProductsByID(ctx, decodedProductID)
	if err != nil {
		return err
	}

	origPrice, discountedPrice, _ := utils.GetOrigAndDiscounted(
		product.IsOnSale,
		product.UnitPriceWithVat,
		product.UnitPriceWithVatCurrency,
		product.SalePriceWithVat,
		product.SalePriceWithVatCurrency,
	)

	existingLines, err := s.dbRO.GetQueries().GetQuotationLinesByQuotationIDAndProductID(ctx, queries.GetQuotationLinesByQuotationIDAndProductIDParams{
		QuotationID: quotation.ID,
		ProductID:   decodedProductID,
	})
	if err != nil {
		return err
	}

	if len(existingLines) > 0 {
		totalQty := quantity
		for _, line := range existingLines {
			totalQty += line.Quantity
		}

		_, err = s.dbRW.GetQueries().UpdateQuotationLineOnAdd(ctx, queries.UpdateQuotationLineOnAddParams{
			Quantity:              totalQty,
			OriginalPriceSnapshot: sql.NullInt64{Valid: true, Int64: origPrice.Amount()},
			SalePriceSnapshot:     sql.NullInt64{Valid: true, Int64: discountedPrice.Amount()},
			Currency:              product.UnitPriceWithoutVatCurrency,
			ID:                    existingLines[0].ID,
		})
		if err != nil {
			return err
		}

		for _, dup := range existingLines[1:] {
			if err := s.dbRW.GetQueries().DeleteQuotationLine(ctx, dup.ID); err != nil {
				return err
			}
		}

		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("quotation_id", s.encoder.Encode(quotation.ID)),
			zap.String("product_id", productID),
			zap.Int64("quantity", totalQty),
			zap.Bool("merged", true),
		)

		return nil
	}

	_, err = s.dbRW.GetQueries().CreateQuotationLine(ctx, queries.CreateQuotationLineParams{
		QuotationID:           quotation.ID,
		ProductID:             decodedProductID,
		Quantity:              quantity,
		OriginalPriceSnapshot: sql.NullInt64{Valid: true, Int64: origPrice.Amount()},
		SalePriceSnapshot:     sql.NullInt64{Valid: true, Int64: discountedPrice.Amount()},
		Currency:              product.UnitPriceWithoutVatCurrency,
	})
	if err != nil {
		return err
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("quotation_id", s.encoder.Encode(quotation.ID)),
		zap.String("product_id", productID),
		zap.Int64("quantity", quantity),
	)

	return nil
}

func (s *QuotationService) RemoveLine(ctx context.Context, lineID string) error {
	const logtag = "[QuotationService] RemoveLine"
	decodedLineID := s.encoder.Decode(lineID)
	if decodedLineID == encode.INVALID {
		return errs.ErrDecode
	}

	if err := s.dbRW.GetQueries().DeleteQuotationLine(ctx, decodedLineID); err != nil {
		return err
	}

	logs.LogCtx(ctx).Info(logtag, zap.String("line_id", lineID))
	return nil
}

func (s *QuotationService) GetLines(ctx context.Context, quotationID string) ([]queries.GetQuotationLinesByQuotationIDRow, error) {
	decodedID := s.encoder.Decode(quotationID)
	if decodedID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	lines, err := s.dbRO.GetQueries().GetQuotationLinesByQuotationID(ctx, decodedID)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func (s *QuotationService) GetSummary(ctx context.Context, quotationID string) (queries.GetQuotationSummaryRow, error) {
	decodedID := s.encoder.Decode(quotationID)
	if decodedID == encode.INVALID {
		return queries.GetQuotationSummaryRow{}, errs.ErrDecode
	}

	summary, err := s.dbRO.GetQueries().GetQuotationSummary(ctx, decodedID)
	if err != nil {
		return queries.GetQuotationSummaryRow{}, err
	}
	return summary, nil
}

func (s *QuotationService) SubmitForReview(ctx context.Context, quotationID string) error {
	const logtag = "[QuotationService] SubmitForReview"
	decodedID := s.encoder.Decode(quotationID)
	if decodedID == encode.INVALID {
		return errs.ErrDecode
	}

	quotation, err := s.dbRO.GetQueries().GetQuotationByID(ctx, decodedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}
		return err
	}

	currentStatus := enums.ParseQuotationStatus(quotation.Status)
	if currentStatus != enums.QUOTATION_STATUS_DRAFT {
		return errs.ErrForbidden
	}

	summary, err := s.GetSummary(ctx, quotationID)
	if err != nil {
		return err
	}
	if summary.TotalItems == 0 {
		return errs.ErrMissingField
	}

	_, err = s.dbRW.GetQueries().UpdateQuotationStatus(ctx, queries.UpdateQuotationStatusParams{
		Status: enums.QUOTATION_STATUS_IN_REVIEW.String(),
		ID:     decodedID,
	})
	if err != nil {
		return err
	}

	if err := s.insertStatusHistory(
		ctx,
		decodedID,
		sql.NullInt64{Valid: false},
		sql.NullString{String: enums.QUOTATION_STATUS_DRAFT.String(), Valid: true},
		enums.QUOTATION_STATUS_IN_REVIEW.String(),
		"Submitted for review",
	); err != nil {
		return err
	}

	logs.LogCtx(ctx).Info(logtag, zap.String("quotation_id", quotationID))
	return nil
}

func (s *QuotationService) ID() string {
	return "Quotation"
}

func (s *QuotationService) Log() {
	logs.Log().Info("[QuotationService] Loaded")
}

var _ IService = (*QuotationService)(nil)
