package checkout

import (
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"context"

	"go.uber.org/zap"
)

func CreateCheckout(
	ctx context.Context,
	dbq *queries.Queries,
	encode encode.IEncode,
	token string,
	checkoutLineProductIDs []string,
) (int64, error) {
	if len(checkoutLineProductIDs) == 0 || token == "" {
		return -1, nil
	}

	if checkoutID, err := dbq.GetCheckoutIDBySessionID(ctx, token); err == nil && checkoutID != 0 {
		return checkoutID, nil
	}

	checkout, err := dbq.CreateCheckout(ctx, token)
	if err != nil {
		return -1, err
	}

	productIDsCount := map[string]int64{}
	for _, checkoutLineProductID := range checkoutLineProductIDs {
		productIDsCount[checkoutLineProductID]++
	}

	created := 0
	for productID, qty := range productIDsCount {
		dbProductID := encode.Decode(productID)
		_, err := dbq.CreateCheckoutLine(
			ctx,
			queries.CreateCheckoutLineParams{
				CheckoutID: checkout.ID,
				ProductID:  dbProductID,
				Quantity:   qty,

				//TODO: (Brandon)
				Name:        "",
				Serial:      "",
				Description: "",
				Amount:      0,
				Currency:    "",
			},
		)
		if err != nil {
			logs.Log().Warn(
				"Can't create checkoutline",
				zap.Error(err),
				zap.Int64("checkout id", checkout.ID),
				zap.String("product id", productID),
			)
		}
		created++
	}

	logs.Log().Info(
		"Created checkout",
		zap.String("token", token),
		zap.Int64("checkout id", checkout.ID),
		zap.Int("checkout lines count", created),
	)

	return checkout.ID, nil
}
