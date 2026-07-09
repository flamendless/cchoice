package errs

import "errors"

var (
	ErrMemo                     = errors.New("[MEMO]: Error on memo service")
	ErrMemoNotFound             = errors.New("[MEMO]: Memo not found")
	ErrMemoGetFailed            = errors.New("[MEMO]: Failed to get memo")
	ErrMemoDeleteFailed         = errors.New("[MEMO]: Failed to delete memo")
	ErrMemoAllFieldsRequired    = errors.New("[MEMO]: All fields are required")
	ErrMemoRecipientsRequired   = errors.New("[MEMO]: At least one staff recipient is required")
	ErrMemoRejectReasonRequired = errors.New("[MEMO]: Reject reason is required")
	ErrMemoDateBeforeToday      = errors.New("[MEMO]: Date must not be before today")
	ErrMemoAlreadyAcknowledged  = errors.New("[MEMO]: Memo already acknowledged")
	ErrMemoSendNotAllowed       = errors.New("[MEMO]: You are not allowed to send emails for this memo")
	ErrMemoEmailRateLimited     = errors.New("[MEMO]: Emails were sent recently. Please wait 24 hours before sending again")
	ErrMemoNoRecipientEmails    = errors.New("[MEMO]: No recipients with email addresses found")
	ErrMemoNotPublished         = errors.New("[MEMO]: Only published memos can send notification emails")
	ErrMemoInvalidStatus        = errors.New("[MEMO]: Invalid status")
)
