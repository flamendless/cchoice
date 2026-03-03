package server

import (
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"strconv"

	"cchoice/internal/types"

	"github.com/alexedwards/scs/v2"
)

const (
	skCheckoutLineProductIDs = "checkout_line_ids"
	skShippingQuotation      = "shipping_quotation"
	skShippingRequest        = "shipping_request"
	skCheckedItems           = "checked_items"
	skLocationLat            = "location_lat"
	skLocationLng            = "location_lng"
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

func GetCheckedItems(ctx context.Context, sm *scs.SessionManager) []string {
	if !sm.Exists(ctx, skCheckedItems) {
		return []string{}
	}

	checkedItems, ok := sm.Get(ctx, skCheckedItems).([]string)
	if !ok {
		return []string{}
	}

	return checkedItems
}

func SetCheckedItems(ctx context.Context, sm *scs.SessionManager, checkedItems []string) {
	sm.Put(ctx, skCheckedItems, checkedItems)
}

func ToggleCheckedItem(ctx context.Context, sm *scs.SessionManager, itemID string) []string {
	checkedItems := GetCheckedItems(ctx, sm)

	for i, id := range checkedItems {
		if id == itemID {
			checkedItems = slices.Delete(checkedItems, i, i+1)
			sm.Put(ctx, skCheckedItems, checkedItems)
			return checkedItems
		}
	}

	checkedItems = append(checkedItems, itemID)
	sm.Put(ctx, skCheckedItems, checkedItems)
	return checkedItems
}

func GetLocation(ctx context.Context, sm *scs.SessionManager) sql.NullString {
	latVal := sm.Get(ctx, skLocationLat)
	lngVal := sm.Get(ctx, skLocationLng)
	if latVal == nil || lngVal == nil {
		return sql.NullString{}
	}
	lat, ok1 := latVal.(float64)
	lng, ok2 := lngVal.(float64)
	if !ok1 || !ok2 {
		return sql.NullString{}
	}
	b, _ := json.Marshal(types.Location{Lat: lat, Lng: lng})
	return sql.NullString{String: string(b), Valid: true}
}

func SetLocation(
	ctx context.Context,
	sm *scs.SessionManager,
	lat, lng string,
) {
	if nLat, errLat := strconv.ParseFloat(lat, 64); errLat == nil {
		if nLng, errLng := strconv.ParseFloat(lng, 64); errLng == nil {
			sm.Put(ctx, skLocationLat, nLat)
			sm.Put(ctx, skLocationLng, nLng)
		}
	}
}
