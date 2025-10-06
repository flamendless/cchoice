package server

import (
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"context"
	"slices"

	"github.com/alexedwards/scs/v2"
)

const (
	skCheckoutLineProductIDs = "checkout_line_ids"
	skShippingQuotation      = "shipping_quotation"
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
		return nil, errs.ErrSessionCheckoutLineProductIDs
	}

	checkoutLineProductIDs = append(checkoutLineProductIDs, productID)
	sm.Put(ctx, skCheckoutLineProductIDs, checkoutLineProductIDs)

	return checkoutLineProductIDs, nil
}

func RemoveFromCheckoutLineProductIDs(
	ctx context.Context,
	sm *scs.SessionManager,
	productID string,
) ([]string, error) {
	if !sm.Exists(ctx, skCheckoutLineProductIDs) {
		logs.Log().Warn("Attempted to remove an item from cart line but cart lines do not exist in session")
		return nil, nil
	}

	checkoutLineProductIDs, ok := sm.Get(ctx, skCheckoutLineProductIDs).([]string)
	if !ok {
		return nil, errs.ErrSessionCheckoutLineProductIDs
	}

	checkoutLineProductIDs = slices.DeleteFunc(checkoutLineProductIDs, func(s string) bool {
		return s == productID
	})

	sm.Put(ctx, skCheckoutLineProductIDs, checkoutLineProductIDs)

	return checkoutLineProductIDs, nil
}
