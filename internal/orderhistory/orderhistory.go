package orderhistory

import (
	"context"
	"database/sql"
	"fmt"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
)

func Record(
	ctx context.Context,
	dbRW database.IService,
	orderID int64,
	staffID sql.NullInt64,
	fromStatus sql.NullString,
	toStatus string,
	notes sql.NullString,
) error {
	return RecordWithQueries(ctx, dbRW.GetQueries(), orderID, staffID, fromStatus, toStatus, notes)
}

func RecordWithQueries(
	ctx context.Context,
	q *queries.Queries,
	orderID int64,
	staffID sql.NullInt64,
	fromStatus sql.NullString,
	toStatus string,
	notes sql.NullString,
) error {
	if _, err := q.CreateOrderStatusHistory(ctx, queries.CreateOrderStatusHistoryParams{
		OrderID:    orderID,
		StaffID:    staffID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Notes:      notes,
	}); err != nil {
		return fmt.Errorf("failed to record order status history: %w", err)
	}
	return nil
}
