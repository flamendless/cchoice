package server

import (
	"cchoice/internal/errs"
	"context"

	"github.com/alexedwards/scs/v2"
)

const (
	skCheckoutLineProductIDs = "checkout_line_ids"
)

func AddToCheckoutLineProductIDs(
	ctx context.Context,
	sm *scs.SessionManager,
	productID string,
) ([]string, error) {
	if !sm.Exists(ctx, skCheckoutLineProductIDs) {
		sm.Put(ctx, skCheckoutLineProductIDs, []string{})
	}

	checkoutLineProductIDs, ok := sm.Get(ctx, skCheckoutLineProductIDs).([]string)
	if !ok {
		return nil, errs.ERR_SESSION_CHECKOUT_LINE_PRODUCT_IDS
	}

	checkoutLineProductIDs = append(checkoutLineProductIDs, productID)
	sm.Put(ctx, skCheckoutLineProductIDs, checkoutLineProductIDs)

	return checkoutLineProductIDs, nil
}
