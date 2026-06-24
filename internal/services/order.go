package services

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
)

type OrderService struct {
	encoder encode.IEncode
	dbRO    database.IService
}

func NewOrderService(
	encoder encode.IEncode,
	dbRO database.IService,
) *OrderService {
	return &OrderService{
		encoder: encoder,
		dbRO:    dbRO,
	}
}

func (s *OrderService) GetForListingAdmin(
	ctx context.Context,
	searchOrderRef string,
	sortBy string,
	sortDir string,
) ([]OrderAdminListItem, error) {
	if sortBy == "" {
		sortBy = "UPDATED_AT"
	}
	if sortDir == "" {
		sortDir = "DESC"
	}

	orders, err := s.dbRO.GetQueries().AdminGetOrdersForListing(ctx, sql.NullString{
		String: searchOrderRef,
		Valid:  searchOrderRef != "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for listing: %w", err)
	}

	sortOrdersForAdmin(orders, sortBy, sortDir)

	result := make([]OrderAdminListItem, 0, len(orders))
	for _, order := range orders {
		result = append(result, OrderAdminListItem{
			ID:             order.ID,
			OrderReference: order.OrderNumber,
			Status:         enums.ParseOrderStatusToEnum(order.Status),
			IsPaid:         order.PaidAt.Valid,
			CreatedAt:      order.CreatedAt.Format(constants.DateTimeLayoutISO),
			UpdatedAt:      order.UpdatedAt.Format(constants.DateTimeLayoutISO),
		})
	}

	return result, nil
}

func (s *OrderService) GetDetailsForAdmin(ctx context.Context, orderID string) (*OrderAdminDetails, error) {
	decoded := s.encoder.Decode(orderID)
	if decoded == encode.INVALID {
		return nil, errs.ErrDecode
	}

	order, err := s.dbRO.GetQueries().AdminGetOrderDetailsByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}

	addressParts := []string{
		order.ShippingAddressLine1,
		order.ShippingAddressLine2,
		order.ShippingCity,
		order.ShippingState,
		order.ShippingPostalCode,
		order.ShippingCountry,
	}
	address := strings.Join(addressParts, ", ")
	address = strings.ReplaceAll(address, ", , ", ", ")
	address = strings.Trim(address, ", ")

	lines, err := s.dbRO.GetQueries().GetOrderLinesByOrderID(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order lines: %w", err)
	}

	lineItems := make([]OrderAdminLineItem, 0, len(lines))
	for _, line := range lines {
		lineItems = append(lineItems, OrderAdminLineItem{
			Name:        line.Name,
			Serial:      line.Serial,
			Description: line.Description,
			UnitPrice:   utils.NewMoney(line.UnitPrice, line.Currency).Display(),
			Quantity:    line.Quantity,
			TotalPrice:  utils.NewMoney(line.TotalPrice, line.Currency).Display(),
		})
	}

	return &OrderAdminDetails{
		Customer: OrderAdminCustomerInfo{
			Name:    order.CustomerName,
			Email:   order.CustomerEmail,
			Phone:   order.CustomerPhone,
			Address: address,
		},
		Lines: lineItems,
	}, nil
}

func (s *OrderService) ID() string {
	return "Order"
}

func (s *OrderService) Log() {
	logs.Log().Info("[OrderService] Loaded")
}

var _ IService = (*OrderService)(nil)

func sortOrdersForAdmin(orders []queries.AdminGetOrdersForListingRow, sortBy, sortDir string) {
	asc := sortDir == "ASC"

	slices.SortFunc(orders, func(a, b queries.AdminGetOrdersForListingRow) int {
		switch sortBy {
		case "CREATED_AT":
			return compareTime(a.CreatedAt, b.CreatedAt, asc)
		case "STATUS":
			return compareString(a.Status, b.Status, asc)
		default:
			return compareTime(a.UpdatedAt, b.UpdatedAt, asc)
		}
	})
}

func compareTime(a, b time.Time, asc bool) int {
	if asc {
		return cmp.Compare(a.Unix(), b.Unix())
	}
	return cmp.Compare(b.Unix(), a.Unix())
}

func compareString(a, b string, asc bool) int {
	if asc {
		return cmp.Compare(a, b)
	}
	return cmp.Compare(b, a)
}
