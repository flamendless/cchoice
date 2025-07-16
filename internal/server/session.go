package server

import (
	"context"
	"errors"

	"github.com/alexedwards/scs/v2"
)

const keyLineItems = "line_items"

func GetOrCreateLineItems(sm *scs.SessionManager, ctx context.Context) ([]string, error) {
	if !sm.Exists(ctx, keyLineItems) {
		sm.Put(ctx, keyLineItems, []string{})
	}

	lineItems, ok := sm.Get(ctx, keyLineItems).([]string)
	if !ok {
		return nil, errors.New("Failed to cast line items to []string")
	}

	return lineItems, nil
}

func UpdateLineItems(
	sm *scs.SessionManager,
	ctx context.Context,
	lineItems []string,
) {
	sm.Put(ctx, keyLineItems, lineItems)
}
