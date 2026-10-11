package errs

import "errors"

var (
	ErrInsufficientLeaveCredits = errors.New("insufficient leave credit hours")
	ErrInvalidLeaveHours        = errors.New("leave hours must be greater than zero")
	ErrTimeOffNotPending        = errors.New("time off request is not pending")
	ErrTimeOffRejectReason      = errors.New("rejection reason is required")
)
