package errs

import "errors"

var (
	ErrTrackedLinkInUseByPromo = errors.New("[TRACKED_LINK]: Cannot delete or unpublish — used by a published promo banner that is active or scheduled")
)
