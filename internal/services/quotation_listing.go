package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func formatQuotationTotal(orig, sale int64, currency any) string {
	curr := utils.StringFromAny(currency)
	origM := utils.NewMoney(orig, curr)
	saleM := utils.NewMoney(sale, curr)
	total, _ := origM.Subtract(saleM)
	return total.Display()
}

func mapAdminQuotationListItem(
	id int64,
	status string,
	createdAt, updatedAt string,
	customerFirst, customerMiddle, customerLast string,
	staffFirst, staffLast sql.NullString,
	totalItems, totalOrig, totalSale int64,
	currency any,
) QuotationAdminListItem {
	assignedTo := "—"
	if staffFirst.Valid || staffLast.Valid {
		parts := make([]string, 0, 2)
		if staffFirst.Valid && staffFirst.String != "" {
			parts = append(parts, staffFirst.String)
		}
		if staffLast.Valid && staffLast.String != "" {
			parts = append(parts, staffLast.String)
		}
		if len(parts) > 0 {
			assignedTo = strings.Join(parts, " ")
		}
	}

	return QuotationAdminListItem{
		ID:           id,
		CustomerName: utils.BuildFullName(customerFirst, customerMiddle, customerLast),
		Status:       enums.ParseQuotationStatus(status),
		AssignedTo:   assignedTo,
		TotalItems:   totalItems,
		TotalDisplay: formatQuotationTotal(totalOrig, totalSale, currency),
		SubmittedAt:  createdAt,
		UpdatedAt:    updatedAt,
	}
}

func mapAdminQuotationFromCreatedAtDesc(rows []queries.AdminGetQuotationsForListingPaginatedCreatedAtDescRow) []QuotationAdminListItem {
	result := make([]QuotationAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminQuotationListItem(
			row.ID,
			row.Status,
			row.CreatedAt.Format(constants.DateTimeLayoutISO),
			row.UpdatedAt.Format(constants.DateTimeLayoutISO),
			row.CustomerFirstName,
			row.CustomerMiddleName.String,
			row.CustomerLastName,
			row.StaffFirstName,
			row.StaffLastName,
			row.TotalItems,
			row.TotalOriginalPrice,
			row.TotalSalePrice,
			row.Currency,
		))
	}
	return result
}

func mapAdminQuotationFromCreatedAtAsc(rows []queries.AdminGetQuotationsForListingPaginatedCreatedAtAscRow) []QuotationAdminListItem {
	result := make([]QuotationAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminQuotationListItem(
			row.ID,
			row.Status,
			row.CreatedAt.Format(constants.DateTimeLayoutISO),
			row.UpdatedAt.Format(constants.DateTimeLayoutISO),
			row.CustomerFirstName,
			row.CustomerMiddleName.String,
			row.CustomerLastName,
			row.StaffFirstName,
			row.StaffLastName,
			row.TotalItems,
			row.TotalOriginalPrice,
			row.TotalSalePrice,
			row.Currency,
		))
	}
	return result
}

func mapAdminQuotationFromStatusDesc(rows []queries.AdminGetQuotationsForListingPaginatedStatusDescRow) []QuotationAdminListItem {
	result := make([]QuotationAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminQuotationListItem(
			row.ID,
			row.Status,
			row.CreatedAt.Format(constants.DateTimeLayoutISO),
			row.UpdatedAt.Format(constants.DateTimeLayoutISO),
			row.CustomerFirstName,
			row.CustomerMiddleName.String,
			row.CustomerLastName,
			row.StaffFirstName,
			row.StaffLastName,
			row.TotalItems,
			row.TotalOriginalPrice,
			row.TotalSalePrice,
			row.Currency,
		))
	}
	return result
}

func mapAdminQuotationFromStatusAsc(rows []queries.AdminGetQuotationsForListingPaginatedStatusAscRow) []QuotationAdminListItem {
	result := make([]QuotationAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminQuotationListItem(
			row.ID,
			row.Status,
			row.CreatedAt.Format(constants.DateTimeLayoutISO),
			row.UpdatedAt.Format(constants.DateTimeLayoutISO),
			row.CustomerFirstName,
			row.CustomerMiddleName.String,
			row.CustomerLastName,
			row.StaffFirstName,
			row.StaffLastName,
			row.TotalItems,
			row.TotalOriginalPrice,
			row.TotalSalePrice,
			row.Currency,
		))
	}
	return result
}

func mapCustomerQuotationListItem(
	id int64,
	status string,
	createdAt string,
	totalItems, totalOrig, totalSale int64,
	currency any,
) QuotationCustomerListItem {
	return QuotationCustomerListItem{
		ID:           id,
		Status:       enums.ParseQuotationStatus(status),
		TotalItems:   totalItems,
		TotalDisplay: formatQuotationTotal(totalOrig, totalSale, currency),
		SubmittedAt:  createdAt,
	}
}

func (s *QuotationService) GetForListingAdminPaginated(
	ctx context.Context,
	search string,
	sortBy string,
	sortDir enums.ListingSortDirection,
	page, perPage int,
) ([]QuotationAdminListItem, int64, int, error) {
	sortBy, sortDir = utils.NormalizeListingSort(sortBy, sortDir, "CREATED_AT")

	searchParam := sql.NullString{
		String: search,
		Valid:  search != "",
	}

	totalCount, err := s.dbRO.GetQueries().AdminCountQuotationsForListing(ctx, searchParam)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count quotations for listing: %w", err)
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	items, err := s.queryQuotationsForAdminListing(ctx, sortBy, sortDir, searchParam, int64(perPage), offset)
	if err != nil {
		return nil, 0, 0, err
	}

	return items, totalCount, page, nil
}

func (s *QuotationService) queryQuotationsForAdminListing(
	ctx context.Context,
	sortBy string,
	sortDir enums.ListingSortDirection,
	search sql.NullString,
	limit, offset int64,
) ([]QuotationAdminListItem, error) {
	q := s.dbRO.GetQueries()

	switch sortBy {
	case "STATUS":
		if sortDir.IsAscending() {
			rows, err := q.AdminGetQuotationsForListingPaginatedStatusAsc(ctx, queries.AdminGetQuotationsForListingPaginatedStatusAscParams{
				Search: search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				return nil, err
			}
			return mapAdminQuotationFromStatusAsc(rows), nil
		}
		rows, err := q.AdminGetQuotationsForListingPaginatedStatusDesc(ctx, queries.AdminGetQuotationsForListingPaginatedStatusDescParams{
			Search: search,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		return mapAdminQuotationFromStatusDesc(rows), nil
	default:
		if sortDir.IsAscending() {
			rows, err := q.AdminGetQuotationsForListingPaginatedCreatedAtAsc(ctx, queries.AdminGetQuotationsForListingPaginatedCreatedAtAscParams{
				Search: search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				return nil, err
			}
			return mapAdminQuotationFromCreatedAtAsc(rows), nil
		}
		rows, err := q.AdminGetQuotationsForListingPaginatedCreatedAtDesc(ctx, queries.AdminGetQuotationsForListingPaginatedCreatedAtDescParams{
			Search: search,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		return mapAdminQuotationFromCreatedAtDesc(rows), nil
	}
}

func (s *QuotationService) GetForListingCustomerPaginated(
	ctx context.Context,
	customerIDStr string,
	sortBy string,
	sortDir enums.ListingSortDirection,
	page, perPage int,
) ([]QuotationCustomerListItem, int64, int, error) {
	decodedCustomerID := s.encoder.Decode(customerIDStr)
	if decodedCustomerID == encode.INVALID {
		return nil, 0, 0, errs.ErrDecode
	}

	sortBy, sortDir = utils.NormalizeListingSort(sortBy, sortDir, "CREATED_AT")

	totalCount, err := s.dbRO.GetQueries().CustomerCountQuotationsForListing(ctx, decodedCustomerID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count customer quotations: %w", err)
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	items, err := s.queryQuotationsForCustomerListing(ctx, decodedCustomerID, sortBy, sortDir, int64(perPage), offset)
	if err != nil {
		return nil, 0, 0, err
	}

	return items, totalCount, page, nil
}

func (s *QuotationService) queryQuotationsForCustomerListing(
	ctx context.Context,
	customerID int64,
	sortBy string,
	sortDir enums.ListingSortDirection,
	limit, offset int64,
) ([]QuotationCustomerListItem, error) {
	q := s.dbRO.GetQueries()

	mapRow := func(id int64, status string, createdAt interface{ Format(string) string }, totalItems, totalOrig, totalSale int64, currency any) QuotationCustomerListItem {
		return mapCustomerQuotationListItem(id, status, createdAt.Format(constants.DateTimeLayoutISO), totalItems, totalOrig, totalSale, currency)
	}

	switch sortBy {
	case "STATUS":
		if sortDir.IsAscending() {
			rows, err := q.CustomerGetQuotationsForListingPaginatedStatusAsc(ctx, queries.CustomerGetQuotationsForListingPaginatedStatusAscParams{
				CustomerID: customerID,
				Limit:      limit,
				Offset:     offset,
			})
			if err != nil {
				return nil, err
			}
			result := make([]QuotationCustomerListItem, 0, len(rows))
			for _, row := range rows {
				result = append(result, mapRow(row.ID, row.Status, row.CreatedAt, row.TotalItems, row.TotalOriginalPrice, row.TotalSalePrice, row.Currency))
			}
			return result, nil
		}
		rows, err := q.CustomerGetQuotationsForListingPaginatedStatusDesc(ctx, queries.CustomerGetQuotationsForListingPaginatedStatusDescParams{
			CustomerID: customerID,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]QuotationCustomerListItem, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapRow(row.ID, row.Status, row.CreatedAt, row.TotalItems, row.TotalOriginalPrice, row.TotalSalePrice, row.Currency))
		}
		return result, nil
	default:
		if sortDir.IsAscending() {
			rows, err := q.CustomerGetQuotationsForListingPaginatedCreatedAtAsc(ctx, queries.CustomerGetQuotationsForListingPaginatedCreatedAtAscParams{
				CustomerID: customerID,
				Limit:      limit,
				Offset:     offset,
			})
			if err != nil {
				return nil, err
			}
			result := make([]QuotationCustomerListItem, 0, len(rows))
			for _, row := range rows {
				result = append(result, mapRow(row.ID, row.Status, row.CreatedAt, row.TotalItems, row.TotalOriginalPrice, row.TotalSalePrice, row.Currency))
			}
			return result, nil
		}
		rows, err := q.CustomerGetQuotationsForListingPaginatedCreatedAtDesc(ctx, queries.CustomerGetQuotationsForListingPaginatedCreatedAtDescParams{
			CustomerID: customerID,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]QuotationCustomerListItem, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapRow(row.ID, row.Status, row.CreatedAt, row.TotalItems, row.TotalOriginalPrice, row.TotalSalePrice, row.Currency))
		}
		return result, nil
	}
}

func (s *QuotationService) mapLineItems(lines []queries.GetQuotationLinesByQuotationIDRow) []QuotationAdminLineItem {
	result := make([]QuotationAdminLineItem, 0, len(lines))
	for _, line := range lines {
		orig := utils.NewMoney(line.OriginalPriceSnapshot.Int64*line.Quantity, line.Currency)
		sale := utils.NewMoney(line.SalePriceSnapshot.Int64*line.Quantity, line.Currency)
		td, _ := orig.Subtract(sale)

		result = append(result, QuotationAdminLineItem{
			BrandName:     line.BrandName,
			ProductSerial: line.ProductSerial,
			Quantity:      line.Quantity,
			TotalPrice:    orig.Display(),
			TotalDiscount: td.Display(),
		})
	}
	return result
}

func (s *QuotationService) GetLinesForAdmin(ctx context.Context, quotationIDStr string) ([]QuotationAdminLineItem, error) {
	lines, err := s.GetLines(ctx, quotationIDStr)
	if err != nil {
		return nil, err
	}
	return s.mapLineItems(lines), nil
}

func mapQuotationStatusHistoryEntry(row queries.GetQuotationStatusHistoryByQuotationIDRow) QuotationStatusHistoryEntry {
	fromStatus := "—"
	if row.FromStatus.Valid && row.FromStatus.String != "" {
		fromStatus = row.FromStatus.String
	}

	staffName := "System"
	if row.StaffID.Valid {
		parts := make([]string, 0, 2)
		if row.StaffFirstName.Valid && row.StaffFirstName.String != "" {
			parts = append(parts, row.StaffFirstName.String)
		}
		if row.StaffLastName.Valid && row.StaffLastName.String != "" {
			parts = append(parts, row.StaffLastName.String)
		}
		if len(parts) > 0 {
			staffName = strings.Join(parts, " ")
		} else {
			staffName = "Staff"
		}
	}

	notes := ""
	if row.Notes.Valid {
		notes = row.Notes.String
	}

	return QuotationStatusHistoryEntry{
		FromStatus: fromStatus,
		ToStatus:   row.ToStatus,
		StaffName:  staffName,
		Notes:      notes,
		CreatedAt:  row.CreatedAt.Format(constants.DateTimeLayoutISO),
	}
}

func buildQuotationFlowSteps(history []QuotationStatusHistoryEntry) []enums.QuotationStatus {
	steps := make([]enums.QuotationStatus, 0, len(history))
	var prev enums.QuotationStatus
	for _, entry := range history {
		status := enums.ParseQuotationStatus(entry.ToStatus)
		if status == enums.QUOTATION_STATUS_UNDEFINED || status == enums.QUOTATION_STATUS_DRAFT {
			continue
		}
		if len(steps) == 0 || status != prev {
			steps = append(steps, status)
			prev = status
		}
	}
	return steps
}

func (s *QuotationService) getStatusHistory(ctx context.Context, quotationID int64) ([]QuotationStatusHistoryEntry, error) {
	rows, err := s.dbRO.GetQueries().GetQuotationStatusHistoryByQuotationID(ctx, quotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotation status history: %w", err)
	}

	history := make([]QuotationStatusHistoryEntry, 0, len(rows))
	for _, row := range rows {
		history = append(history, mapQuotationStatusHistoryEntry(row))
	}
	return history, nil
}

func (s *QuotationService) GetStatusHistoryForAdmin(ctx context.Context, quotationIDStr string) (*QuotationAdminTrackData, error) {
	decoded := s.encoder.Decode(quotationIDStr)
	if decoded == encode.INVALID {
		return nil, errs.ErrDecode
	}

	quotation, err := s.dbRO.GetQueries().GetQuotationByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get quotation: %w", err)
	}

	history, err := s.getStatusHistory(ctx, decoded)
	if err != nil {
		return nil, err
	}

	currentStatus := enums.ParseQuotationStatus(quotation.Status)

	return &QuotationAdminTrackData{
		ID:            quotation.ID,
		CurrentStatus: currentStatus,
		History:       history,
		FlowSteps:     buildQuotationFlowSteps(history),
	}, nil
}

func (s *QuotationService) GetDetailForCustomer(ctx context.Context, customerIDStr, quotationIDStr string) (*QuotationCustomerDetailData, error) {
	decodedCustomerID := s.encoder.Decode(customerIDStr)
	if decodedCustomerID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	decodedQuotationID := s.encoder.Decode(quotationIDStr)
	if decodedQuotationID == encode.INVALID {
		return nil, errs.ErrDecode
	}

	quotation, err := s.dbRO.GetQueries().GetQuotationByID(ctx, decodedQuotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get quotation: %w", err)
	}

	if quotation.CustomerID != decodedCustomerID {
		return nil, errs.ErrForbidden
	}

	if enums.ParseQuotationStatus(quotation.Status) == enums.QUOTATION_STATUS_DRAFT {
		return nil, errs.ErrNotFound
	}

	lines, err := s.GetLines(ctx, quotationIDStr)
	if err != nil {
		return nil, err
	}

	summary, err := s.GetSummary(ctx, quotationIDStr)
	if err != nil {
		return nil, err
	}

	summaryModel := models.QuotationSummary{}
	if summary.TotalItems > 0 {
		summaryModel.TotalItems = summary.TotalItems

		var orig, discount int64
		if originalPrice, ok := summary.TotalOriginalPrice.(int64); ok {
			orig = originalPrice
		}
		if discounts, ok := summary.TotalSalePrice.(int64); ok {
			discount = discounts
		}

		origM := utils.NewMoney(orig, summary.Currency)
		discountM := utils.NewMoney(discount, summary.Currency)
		total, _ := origM.Subtract(discountM)
		summaryModel.TotalPrice = origM.Display()
		summaryModel.TotalDiscounts = discountM.Display()
		summaryModel.Total = total.Display()
	}

	history, err := s.getStatusHistory(ctx, decodedQuotationID)
	if err != nil {
		return nil, err
	}

	currentStatus := enums.ParseQuotationStatus(quotation.Status)

	return &QuotationCustomerDetailData{
		ID:             quotation.ID,
		Status:         currentStatus,
		SubmittedAt:    quotation.CreatedAt.Format(constants.DateTimeLayoutISO),
		UpdatedAt:      quotation.UpdatedAt.Format(constants.DateTimeLayoutISO),
		Lines:          s.mapLineItems(lines),
		TotalItems:     summaryModel.TotalItems,
		TotalPrice:     summaryModel.TotalPrice,
		TotalDiscounts: summaryModel.TotalDiscounts,
		Total:          summaryModel.Total,
		Track: QuotationAdminTrackData{
			ID:            quotation.ID,
			CurrentStatus: currentStatus,
			History:       history,
			FlowSteps:     buildQuotationFlowSteps(history),
		},
	}, nil
}

func (s *QuotationService) insertStatusHistory(
	ctx context.Context,
	quotationID int64,
	staffID sql.NullInt64,
	fromStatus sql.NullString,
	toStatus string,
	notes string,
) error {
	status := enums.ParseQuotationStatus(toStatus)
	if status == enums.QUOTATION_STATUS_UNDEFINED {
		return errs.ErrEnumInvalid
	}

	notesParam := sql.NullString{String: notes, Valid: notes != ""}

	_, err := s.dbRW.GetQueries().CreateQuotationStatusHistory(ctx, queries.CreateQuotationStatusHistoryParams{
		QuotationID: quotationID,
		StaffID:     staffID,
		FromStatus:  fromStatus,
		ToStatus:    toStatus,
		Notes:       notesParam,
	})
	return err
}

func (s *QuotationService) Approve(
	ctx context.Context,
	actingStaffIDStr string,
	quotationIDStr string,
	assignedStaffIDStr string,
	notes string,
) error {
	const logtag = "[QuotationService] Approve"
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, actingStaffIDStr, constants.ActionApprove, constants.ModuleQuotations, result, nil); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}()

	decodedQuotationID := s.encoder.Decode(quotationIDStr)
	if decodedQuotationID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	assignedStaffID := s.encoder.Decode(assignedStaffIDStr)
	if assignedStaffID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrMissingField
	}

	actingStaffID := s.encoder.Decode(actingStaffIDStr)
	if actingStaffID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}

	quotation, err := s.dbRO.GetQueries().GetQuotationByID(ctx, decodedQuotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = errs.ErrNotFound.Error()
			return errs.ErrNotFound
		}
		result = err.Error()
		return fmt.Errorf("failed to get quotation: %w", err)
	}

	currentStatus := enums.ParseQuotationStatus(quotation.Status)
	if currentStatus != enums.QUOTATION_STATUS_IN_REVIEW {
		result = errs.ErrQuotationNotApprovable.Error()
		return errs.ErrQuotationNotApprovable
	}

	if _, err := s.dbRO.GetQueries().GetStaffByID(ctx, assignedStaffID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			result = errs.ErrNotFound.Error()
			return errs.ErrNotFound
		}
		result = err.Error()
		return fmt.Errorf("failed to get assigned staff: %w", err)
	}

	_, err = s.dbRW.GetQueries().ApproveQuotation(ctx, queries.ApproveQuotationParams{
		Status:                enums.QUOTATION_STATUS_APPROVED.String(),
		AcknowledgedByStaffID: sql.NullInt64{Int64: assignedStaffID, Valid: true},
		ID:                    decodedQuotationID,
	})
	if err != nil {
		result = err.Error()
		return err
	}

	notesTrimmed := strings.TrimSpace(notes)
	if err := s.insertStatusHistory(
		ctx,
		decodedQuotationID,
		sql.NullInt64{Int64: actingStaffID, Valid: true},
		sql.NullString{String: currentStatus.String(), Valid: true},
		enums.QUOTATION_STATUS_APPROVED.String(),
		notesTrimmed,
	); err != nil {
		result = err.Error()
		return err
	}

	logs.LogCtx(ctx).Info(logtag, zap.String("quotation_id", quotationIDStr))
	return nil
}
