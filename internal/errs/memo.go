package errs

import "errors"

var (
	ErrMemo                     = errors.New("[MEMO]: Error on memo service")
	ErrMemoNotFound             = errors.New("[MEMO]: Memo not found")
	ErrMemoRecipientsRequired   = errors.New("[MEMO]: At least one staff recipient is required")
	ErrMemoRejectReasonRequired = errors.New("[MEMO]: Reject reason is required")
	ErrMemoDateBeforeToday      = errors.New("[MEMO]: Date must not be before today")
	ErrMemoAlreadyAcknowledged  = errors.New("[MEMO]: Memo already acknowledged")
)
