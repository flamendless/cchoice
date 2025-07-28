package cart

import (
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

func CreateCart(
	ctx context.Context,
	dbq *queries.Queries,
	encode encode.IEncode,
	token string,
	checkoutLineProductIDs []string,
) (int64, error) {
	if len(checkoutLineProductIDs) == 0 || token == "" {
		return -1, nil
	}

	var checkoutID int64 = -1

	existingCheckoutID, err := dbq.GetCheckoutIDBySessionID(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			checkout, err := dbq.CreateCheckout(ctx, token)
			if err != nil {
				return -1, err
			}
			checkoutID = checkout.ID
		}
	} else {
		checkoutID = existingCheckoutID
	}

	productIDsCount := map[string]int64{}
	for _, checkoutLineProductID := range checkoutLineProductIDs {
		productIDsCount[checkoutLineProductID]++
	}

	created := 0
	for productID, qty := range productIDsCount {
		dbProductID := encode.Decode(productID)
		exists, err := dbq.CheckCheckoutLineExistsByCheckoutIDAndProductID(
			ctx,
			queries.CheckCheckoutLineExistsByCheckoutIDAndProductIDParams{
				CheckoutID: checkoutID,
				ProductID: dbProductID,
			},
		)
		if err != nil || exists == 1 {
			continue
		}

		if _, err := dbq.CreateCheckoutLine(
			ctx,
			queries.CreateCheckoutLineParams{
				CheckoutID: checkoutID,
				ProductID:  dbProductID,
				Quantity:   qty,

				//TODO: (Brandon)
				Name:        "",
				Serial:      "",
				Description: "",
				Amount:      0,
				Currency:    "",
			},
		); err != nil {
			logs.Log().Warn(
				"Can't create checkoutline",
				zap.Error(err),
				zap.Int64("checkout id", checkoutID),
				zap.String("product id", productID),
			)
		}
		created++
	}

	logs.Log().Info(
		"Created checkout",
		zap.String("token", token),
		zap.Int64("checkout id", checkoutID),
		zap.Int("checkout lines count", created),
	)

	return checkoutID, nil
}

func GetCheckoutLines(
	ctx context.Context,
	dbRO database.Service,
	token string,
) ([]queries.GetCheckoutLinesByCheckoutIDRow, error) {
	checkoutID, err := dbRO.GetQueries().GetCheckoutIDBySessionID(ctx, token)
	if err != nil {
		return nil, err
	}

	checkoutLines, err := dbRO.GetQueries().GetCheckoutLinesByCheckoutID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	return checkoutLines, nil
}
