package errs

import "errors"

var (
	ErrHoliday             = errors.New("[HOLIDAY]: Error on holiday service")
	ErrHolidayNotFound     = errors.New("[HOLIDAY]: Holiday not found")
	ErrHolidayAlreadyExist = errors.New("[HOLIDAY]: Holiday already exists")
	ErrHolidayInvalidDate  = errors.New("[HOLIDAY]: Invalid date")
)
