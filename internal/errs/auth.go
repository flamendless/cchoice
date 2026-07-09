package errs

import "errors"

var (
	ErrInvalidEmailOrPasswordFormat = errors.New("[AUTH]: Invalid email or password format")
	ErrInvalidEmailOrPassword       = errors.New("[AUTH]: Invalid email or password")
	ErrLoginRequired                = errors.New("[AUTH]: Login to access page")
	ErrLogInFirst                   = errors.New("[AUTH]: Log in first")
	ErrPasswordsDoNotMatch          = errors.New("[AUTH]: Passwords do not match")
	ErrUnableToLoadProfile          = errors.New("[AUTH]: Unable to load profile")
)
