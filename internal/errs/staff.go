package errs

import "errors"

var (
	ErrStaffResigned             = errors.New("[STAFF]: account is resigned")
	ErrStaffIDRequired           = errors.New("[STAFF]: staff id is required")
	ErrStaffNotFound             = errors.New("[STAFF]: Employee not found")
	ErrStaffLoadFailed           = errors.New("[STAFF]: Failed to load employee")
	ErrStaffUpdateFailed         = errors.New("[STAFF]: Failed to update profile")
	ErrStaffPasswordUpdateFailed = errors.New("[STAFF]: Failed to update password")
	ErrAttendanceNoData          = errors.New("[STAFF]: no attendance data found, skipping report generation")
)
