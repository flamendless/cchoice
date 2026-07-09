package errs

import "errors"

var (
	ErrHoliday                  = errors.New("[HOLIDAY]: Error on holiday service")
	ErrHolidayNotFound          = errors.New("[HOLIDAY]: Holiday not found")
	ErrHolidayAlreadyExist      = errors.New("[HOLIDAY]: Holiday already exists")
	ErrHolidayInvalidDate       = errors.New("[HOLIDAY]: Invalid date")
	ErrHolidayFieldsRequired    = errors.New("[HOLIDAY]: date, name, and type are required")
	ErrHolidayInvalidType       = errors.New("[HOLIDAY]: Invalid holiday type")
	ErrHolidayCreateFailed      = errors.New("[HOLIDAY]: Failed to create holiday")
	ErrHolidayUpdateFailed      = errors.New("[HOLIDAY]: Failed to update holiday")
	ErrHolidayDeleteFailed      = errors.New("[HOLIDAY]: Failed to delete holiday")
)
