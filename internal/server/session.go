package server

import (
	"context"
	"errors"

	"github.com/alexedwards/scs/v2"
)

const sessionKeyProductIDs = "product_ids"

func AddProductID(
	sm *scs.SessionManager,
	ctx context.Context,
	productID string,
) ([]string, error) {
	if !sm.Exists(ctx, sessionKeyProductIDs) {
		sm.Put(ctx, sessionKeyProductIDs, []string{})
	}

	productIDs, ok := sm.Get(ctx, sessionKeyProductIDs).([]string)
	if !ok {
		return nil, errors.New("Failed to cast product IDs to []string")
	}

	//TODO: (Brandon) - there should be a max number of items in a cart
	productIDs = append(productIDs, productID)
	sm.Put(ctx, sessionKeyProductIDs, productIDs)

	return productIDs, nil
}
