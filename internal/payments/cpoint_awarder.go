package payments

import (
	"context"

	"cchoice/internal/database/queries"
)

type ICPointAwarder interface {
	AwardForPaidOrder(ctx context.Context, qtx *queries.Queries, order queries.TblOrder) (int64, error)
}
