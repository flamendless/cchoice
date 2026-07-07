package services

import (
	"database/sql"
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
)

func normalizeOrderListingSort(sortBy, sortDir string) (string, string) {
	if sortBy == "" {
		sortBy = "UPDATED_AT"
	}
	if sortDir == "" {
		sortDir = "DESC"
	}
	return sortBy, sortDir
}

func formatAdminOrderAddress(parts ...string) string {
	address := strings.Join(parts, ", ")
	address = strings.ReplaceAll(address, ", , ", ", ")
	return strings.Trim(address, ", ")
}

func mapAdminOrderAddressInfo(
	line1, line2, city, state, postalCode, country string,
	formattedAddress sql.NullString,
) OrderAdminAddressInfo {
	address := OrderAdminAddressInfo{
		Line1:      line1,
		Line2:      line2,
		City:       city,
		State:      state,
		PostalCode: postalCode,
		Country:    country,
	}
	if formattedAddress.Valid && formattedAddress.String != "" {
		address.FormattedAddress = formattedAddress.String
		return address
	}
	address.FormattedAddress = formatAdminOrderAddress(line1, line2, city, state, postalCode, country)
	return address
}

func formatAdminOrderNullableString(value sql.NullString) string {
	if !value.Valid || value.String == "" {
		return "-"
	}
	return value.String
}

func formatAdminOrderPaymentPaidAt(value sql.NullTime) string {
	if !value.Valid || value.Time.IsZero() {
		return "-"
	}
	return value.Time.Format(constants.DateTimeLayoutISO)
}

func formatAdminOrderPaidAt(orderPaidAt, paymentPaidAt sql.NullTime) string {
	if orderPaidAt.Valid {
		return orderPaidAt.Time.Format(constants.DateTimeLayoutISO)
	}
	return formatAdminOrderPaymentPaidAt(paymentPaidAt)
}

func formatAdminOrderPaymentAmount(amount sql.NullInt64, currency string) string {
	if !amount.Valid {
		return "-"
	}
	return utils.NewMoney(amount.Int64, currency).Display()
}

func resolveOrderLineThumbnailURL(cdnThumbnail sql.NullString, thumbnailPath string) string {
	if cdnThumbnail.Valid && cdnThumbnail.String != "" {
		return cdnThumbnail.String
	}
	if thumbnailPath != "" && thumbnailPath != constants.PathEmptyImage {
		return utils.URL(constants.ToPath1280(thumbnailPath))
	}
	return utils.URL(constants.PathEmptyImage)
}

func mapAdminOrderListItemFromRow(
	id int64,
	orderNumber string,
	status string,
	paidAt sql.NullTime,
	createdAt, updatedAt time.Time,
	earnedCPoints int64,
) OrderAdminListItem {
	return OrderAdminListItem{
		ID:             id,
		OrderReference: orderNumber,
		Status:         enums.ParseOrderStatusToEnum(status),
		IsPaid:         paidAt.Valid,
		CreatedAt:      createdAt.Format(constants.DateTimeLayoutISO),
		UpdatedAt:      updatedAt.Format(constants.DateTimeLayoutISO),
		EarnedCPoints:  earnedCPoints,
	}
}

func mapAdminOrderListItemsFromUpdatedAtDesc(rows []queries.AdminGetOrdersForListingPaginatedUpdatedAtDescRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}

func mapAdminOrderListItemsFromUpdatedAtAsc(rows []queries.AdminGetOrdersForListingPaginatedUpdatedAtAscRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}

func mapAdminOrderListItemsFromCreatedAtDesc(rows []queries.AdminGetOrdersForListingPaginatedCreatedAtDescRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}

func mapAdminOrderListItemsFromCreatedAtAsc(rows []queries.AdminGetOrdersForListingPaginatedCreatedAtAscRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}

func mapAdminOrderListItemsFromStatusDesc(rows []queries.AdminGetOrdersForListingPaginatedStatusDescRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}

func mapAdminOrderListItemsFromStatusAsc(rows []queries.AdminGetOrdersForListingPaginatedStatusAscRow) []OrderAdminListItem {
	result := make([]OrderAdminListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAdminOrderListItemFromRow(row.ID, row.OrderNumber, row.Status, row.PaidAt, row.CreatedAt, row.UpdatedAt, row.EarnedCpoints))
	}
	return result
}
