package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *CPointService) AwardForPaidOrder(ctx context.Context, qtx *queries.Queries, order queries.TblOrder) (int64, error) {
	const logtag = "[CPointService] AwardForPaidOrder"

	earnedCPoints := utils.CalculateOrderEarnedCPoints(order.TotalAmount)
	if earnedCPoints <= 0 || !order.CustomerID.Valid {
		return earnedCPoints, nil
	}

	oneYearLater := time.Now().AddDate(1, 0, 0)
	code := s.GenerateCode()

	if _, err := qtx.CreateRedeemedCpoint(ctx, queries.CreateRedeemedCpointParams{
		CustomerID:  order.CustomerID.Int64,
		Code:        code,
		Value:       earnedCPoints,
		ProductSkus: sql.NullString{},
		ExpiresAt:   sql.NullString{String: oneYearLater.Format(time.RFC3339), Valid: true},
	}); err != nil {
		return 0, fmt.Errorf("failed to create order reward cpoint: %w", err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("order_id", order.ID),
		zap.Int64("customer_id", order.CustomerID.Int64),
		zap.Int64("earned_cpoints", earnedCPoints),
		zap.String("code", code),
	)

	return earnedCPoints, nil
}

var _ payments.ICPointAwarder = (*CPointService)(nil)
