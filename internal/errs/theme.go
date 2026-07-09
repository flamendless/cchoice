package errs

import "errors"

var (
	ErrTheme                 = errors.New("[THEME]: Error on theme service")
	ErrThemeNotFound         = errors.New("[THEME]: Theme not found")
	ErrThemeOverlappingDates = errors.New("[THEME]: Theme date range overlaps with an existing theme")
	ErrThemeInvalidTitle     = errors.New("[THEME]: Title must be alphanumeric and at most 24 characters")
	ErrThemeInvalidConfig    = errors.New("[THEME]: Invalid theme configuration for the given configuration type")
	ErrThemePastDate         = errors.New("[THEME]: Start and end dates must not be in the past")
)
