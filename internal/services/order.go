package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
)

type OrderService struct {
	encoder     encode.IEncode
	dbRO        database.IService
	dbRW        database.IService
	staffLog    *StaffLogsService
	emailRunner *jobs.EmailJobRunner
}

func NewOrderService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
	emailRunner *jobs.EmailJobRunner,
) *OrderService {
	return &OrderService{
		encoder:     encoder,
		dbRO:        dbRO,
		dbRW:        dbRW,
		staffLog:    staffLog,
		emailRunner: emailRunner,
	}
}

func (s *OrderService) GetForListingAdminPaginated(
	ctx context.Context,
	searchOrderRef string,
	sortBy string,
	sortDir string,
	page, perPage int,
) ([]OrderAdminListItem, int64, int, error) {
	sortBy, sortDir = normalizeOrderListingSort(sortBy, sortDir)

	searchParam := sql.NullString{
		String: searchOrderRef,
		Valid:  searchOrderRef != "",
	}

	totalCount, err := s.dbRO.GetQueries().AdminCountOrdersForListing(ctx, searchParam)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count orders for listing: %w", err)
	}

	page = models.ClampPage(page, totalCount, perPage)
	offset := int64((page - 1) * perPage)

	orders, err := s.queryOrdersForListingPaginated(ctx, sortBy, sortDir, searchParam, int64(perPage), offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get orders for listing: %w", err)
	}

	return orders, totalCount, page, nil
}

func (s *OrderService) queryOrdersForListingPaginated(
	ctx context.Context,
	sortBy, sortDir string,
	searchOrderRef sql.NullString,
	limit, offset int64,
) ([]OrderAdminListItem, error) {
	q := s.dbRO.GetQueries()

	switch sortBy {
	case "CREATED_AT":
		if sortDir == "ASC" {
			rows, err := q.AdminGetOrdersForListingPaginatedCreatedAtAsc(ctx, queries.AdminGetOrdersForListingPaginatedCreatedAtAscParams{
				SearchOrderRef: searchOrderRef,
				Limit:          limit,
				Offset:         offset,
			})
			if err != nil {
				return nil, err
			}
			return mapAdminOrderListItemsFromCreatedAtAsc(rows), nil
		}
		rows, err := q.AdminGetOrdersForListingPaginatedCreatedAtDesc(ctx, queries.AdminGetOrdersForListingPaginatedCreatedAtDescParams{
			SearchOrderRef: searchOrderRef,
			Limit:          limit,
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		return mapAdminOrderListItemsFromCreatedAtDesc(rows), nil
	case "STATUS":
		if sortDir == "ASC" {
			rows, err := q.AdminGetOrdersForListingPaginatedStatusAsc(ctx, queries.AdminGetOrdersForListingPaginatedStatusAscParams{
				SearchOrderRef: searchOrderRef,
				Limit:          limit,
				Offset:         offset,
			})
			if err != nil {
				return nil, err
			}
			return mapAdminOrderListItemsFromStatusAsc(rows), nil
		}
		rows, err := q.AdminGetOrdersForListingPaginatedStatusDesc(ctx, queries.AdminGetOrdersForListingPaginatedStatusDescParams{
			SearchOrderRef: searchOrderRef,
			Limit:          limit,
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		return mapAdminOrderListItemsFromStatusDesc(rows), nil
	default:
		if sortDir == "ASC" {
			rows, err := q.AdminGetOrdersForListingPaginatedUpdatedAtAsc(ctx, queries.AdminGetOrdersForListingPaginatedUpdatedAtAscParams{
				SearchOrderRef: searchOrderRef,
				Limit:          limit,
				Offset:         offset,
			})
			if err != nil {
				return nil, err
			}
			return mapAdminOrderListItemsFromUpdatedAtAsc(rows), nil
		}
		rows, err := q.AdminGetOrdersForListingPaginatedUpdatedAtDesc(ctx, queries.AdminGetOrdersForListingPaginatedUpdatedAtDescParams{
			SearchOrderRef: searchOrderRef,
			Limit:          limit,
			Offset:         offset,
		})
		if err != nil {
			return nil, err
		}
		return mapAdminOrderListItemsFromUpdatedAtDesc(rows), nil
	}
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

	lines, err := s.dbRO.GetQueries().AdminGetOrderLinesByOrderID(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order lines: %w", err)
	}

	lineItems := make([]OrderAdminLineItem, 0, len(lines))
	for _, line := range lines {
		lineItems = append(lineItems, OrderAdminLineItem{
			ThumbnailURL: resolveOrderLineThumbnailURL(line.CdnUrlThumbnail, line.ThumbnailPath),
			Name:         line.Name,
			Serial:       line.Serial,
			UnitPrice:    utils.NewMoney(line.UnitPrice, line.Currency).Display(),
			Quantity:     line.Quantity,
			TotalPrice:   utils.NewMoney(line.TotalPrice, line.Currency).Display(),
		})
	}

	currency := order.Currency
	return &OrderAdminDetails{
		Order: OrderAdminInfo{
			OrderReference: order.OrderNumber,
			Status:         enums.ParseOrderStatusToEnum(order.Status),
			Notes:          formatAdminOrderNullableString(order.Notes),
			Remarks:        formatAdminOrderNullableString(order.Remarks),
			CreatedAt:      order.CreatedAt.Format(constants.DateTimeLayoutISO),
			UpdatedAt:      order.UpdatedAt.Format(constants.DateTimeLayoutISO),
			EarnedCPoints:  utils.FormatEarnedCPoints(order.EarnedCpoints),
		},
		Payment: OrderAdminPaymentInfo{
			Gateway:         formatAdminOrderNullableString(order.PaymentGateway),
			Status:          formatAdminOrderNullableString(order.PaymentStatus),
			ReferenceNumber: formatAdminOrderNullableString(order.PaymentReferenceNumber),
			PaymentMethod:   formatAdminOrderNullableString(order.PaymentMethodType),
			TotalAmount:     formatAdminOrderPaymentAmount(order.PaymentTotalAmount, currency),
			PaidAt:          formatAdminOrderPaidAt(order.PaidAt, order.PaymentPaidAt),
			Description:     formatAdminOrderNullableString(order.PaymentDescription),
			MetadataNotes:   formatAdminOrderNullableString(order.PaymentMetadataNotes),
			MetadataRemarks: formatAdminOrderNullableString(order.PaymentMetadataRemarks),
			CustomerNumber:  formatAdminOrderNullableString(order.PaymentMetadataCustomerNumber),
		},
		Shipping: OrderAdminShippingInfo{
			OrderAdminAddressInfo: mapAdminOrderAddressInfo(
				order.ShippingAddressLine1,
				order.ShippingAddressLine2,
				order.ShippingCity,
				order.ShippingState,
				order.ShippingPostalCode,
				order.ShippingCountry,
				order.ShippingFormattedAddress,
			),
			Service:        formatAdminOrderNullableString(order.ShippingService),
			OrderID:        formatAdminOrderNullableString(order.ShippingOrderID),
			TrackingNumber: formatAdminOrderNullableString(order.ShippingTrackingNumber),
			ETA:            formatAdminOrderNullableString(order.ShippingEta),
		},
		Billing: mapAdminOrderAddressInfo(
			order.BillingAddressLine1,
			order.BillingAddressLine2,
			order.BillingCity,
			order.BillingState,
			order.BillingPostalCode,
			order.BillingCountry,
			order.BillingFormattedAddress,
		),
		Customer: OrderAdminCustomerInfo{
			Name:  order.CustomerName,
			Email: order.CustomerEmail,
			Phone: order.CustomerPhone,
		},
		Summary: OrderAdminAmountSummary{
			Subtotal: utils.NewMoney(order.SubtotalAmount, currency).Display(),
			Shipping: utils.NewMoney(order.ShippingAmount, currency).Display(),
			Discount: utils.NewMoney(order.DiscountAmount, currency).Display(),
			Total:    utils.NewMoney(order.TotalAmount, currency).Display(),
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
