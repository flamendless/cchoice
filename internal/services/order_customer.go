package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

func normalizeCustomerOrderListingSort(sortBy, sortDir string) (string, string) {
	if sortBy == "" {
		sortBy = "CREATED_AT"
	}
	if sortDir == "" {
		sortDir = "DESC"
	}
	return sortBy, sortDir
}

func mapCustomerOrderListItemFromRow(
	id int64,
	orderNumber string,
	status string,
	paidAt sql.NullTime,
	createdAt time.Time,
	earnedCPoints int64,
) OrderCustomerListItem {
	return OrderCustomerListItem{
		ID:             id,
		OrderReference: orderNumber,
		Status:         enums.ParseOrderStatusToEnum(status),
		IsPaid:         paidAt.Valid,
		OrderedAt:      createdAt.Format(constants.DateTimeLayoutISO),
		EarnedCPoints:  earnedCPoints,
	}
}

func mapCustomerOrderListItemsFromCreatedAtDesc(rows []queries.CustomerGetOrdersForListingPaginatedCreatedAtDescRow) []OrderCustomerListItem {
	result := make([]OrderCustomerListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCustomerOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.EarnedCpoints))
	}
	return result
}

func mapCustomerOrderListItemsFromCreatedAtAsc(rows []queries.CustomerGetOrdersForListingPaginatedCreatedAtAscRow) []OrderCustomerListItem {
	result := make([]OrderCustomerListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCustomerOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.EarnedCpoints))
	}
	return result
}

func mapCustomerOrderListItemsFromStatusDesc(rows []queries.CustomerGetOrdersForListingPaginatedStatusDescRow) []OrderCustomerListItem {
	result := make([]OrderCustomerListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCustomerOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.EarnedCpoints))
	}
	return result
}

func mapCustomerOrderListItemsFromStatusAsc(rows []queries.CustomerGetOrdersForListingPaginatedStatusAscRow) []OrderCustomerListItem {
	result := make([]OrderCustomerListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCustomerOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.EarnedCpoints))
	}
	return result
}

func (s *OrderService) verifyCustomerOwnsOrder(ctx context.Context, customerIDStr, orderIDStr string) error {
	customerDecoded := s.encoder.Decode(customerIDStr)
	if customerDecoded == encode.INVALID {
		return errs.ErrDecode
	}

	orderDecoded := s.encoder.Decode(orderIDStr)
	if orderDecoded == encode.INVALID {
		return errs.ErrDecode
	}

	order, err := s.dbRO.GetQueries().GetOrderByID(ctx, orderDecoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !order.CustomerID.Valid || order.CustomerID.Int64 != customerDecoded {
		return errs.ErrNotFound
	}

	return nil
}

func (s *OrderService) GetForListingCustomerPaginated(
	ctx context.Context,
	customerIDStr string,
	searchOrderRef string,
	sortBy string,
	sortDir string,
	page, perPage int,
) ([]OrderCustomerListItem, int64, int, error) {
	customerDecoded := s.encoder.Decode(customerIDStr)
	if customerDecoded == encode.INVALID {
		return nil, 0, 0, errs.ErrDecode
	}

	sortBy, sortDir = normalizeCustomerOrderListingSort(sortBy, sortDir)

	searchParam := sql.NullString{
		String: searchOrderRef,
		Valid:  searchOrderRef != "",
	}

	countParams := queries.CustomerCountOrdersForListingParams{
		CustomerID:     sql.NullInt64{Int64: customerDecoded, Valid: true},
		SearchOrderRef: searchParam,
	}

	totalCount, err := s.dbRO.GetQueries().CustomerCountOrdersForListing(ctx, countParams)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count customer orders for listing: %w", err)
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	orders, err := s.queryCustomerOrdersForListingPaginated(ctx, customerDecoded, sortBy, sortDir, searchParam, int64(perPage), offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get customer orders for listing: %w", err)
	}

	return orders, totalCount, page, nil
}

func (s *OrderService) queryCustomerOrdersForListingPaginated(
	ctx context.Context,
	customerID int64,
	sortBy, sortDir string,
	searchOrderRef sql.NullString,
	limit, offset int64,
) ([]OrderCustomerListItem, error) {
	q := s.dbRO.GetQueries()
	customerParam := sql.NullInt64{Int64: customerID, Valid: true}

	switch sortBy {
	case "STATUS":
		if sortDir == "ASC" {
			rows, err := q.CustomerGetOrdersForListingPaginatedStatusAsc(ctx, queries.CustomerGetOrdersForListingPaginatedStatusAscParams{
				CustomerID:     customerParam,
				SearchOrderRef: searchOrderRef,
				Limit:          limit,
				Offset:         offset,
			})
			if err != nil {
				return nil, err
			}
			return mapCustomerOrderListItemsFromStatusAsc(rows), nil
		}
		rows, err := q.CustomerGetOrdersForListingPaginatedStatusDesc(ctx, queries.CustomerGetOrdersForListingPaginatedStatusDescParams{
			CustomerID:     customerParam,
			SearchOrderRef: searchOrderRef,
			Limit:          limit,
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		return mapCustomerOrderListItemsFromStatusDesc(rows), nil
	default:
		if sortDir == "ASC" {
			rows, err := q.CustomerGetOrdersForListingPaginatedCreatedAtAsc(ctx, queries.CustomerGetOrdersForListingPaginatedCreatedAtAscParams{
				CustomerID:     customerParam,
				SearchOrderRef: searchOrderRef,
				Limit:          limit,
				Offset:         offset,
			})
			if err != nil {
				return nil, err
			}
			return mapCustomerOrderListItemsFromCreatedAtAsc(rows), nil
		}
		rows, err := q.CustomerGetOrdersForListingPaginatedCreatedAtDesc(ctx, queries.CustomerGetOrdersForListingPaginatedCreatedAtDescParams{
			CustomerID:     customerParam,
			SearchOrderRef: searchOrderRef,
			Limit:          limit,
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		return mapCustomerOrderListItemsFromCreatedAtDesc(rows), nil
	}
}

func (s *OrderService) GetDetailsForCustomer(ctx context.Context, customerIDStr, orderIDStr string) (*OrderAdminDetails, error) {
	if err := s.verifyCustomerOwnsOrder(ctx, customerIDStr, orderIDStr); err != nil {
		return nil, err
	}
	return s.GetDetailsForAdmin(ctx, orderIDStr)
}

func (s *OrderService) GetStatusHistoryForCustomer(ctx context.Context, customerIDStr, orderIDStr string) (*OrderAdminTrackData, error) {
	if err := s.verifyCustomerOwnsOrder(ctx, customerIDStr, orderIDStr); err != nil {
		return nil, err
	}
	return s.GetStatusHistoryForAdmin(ctx, orderIDStr)
}
