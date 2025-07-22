package errs

import "errors"

var (
	ErrDecode = errors.New("[ENCODE]: Error in decoding")
	ErrEncode = errors.New("[ENCODE]: Error in encoding")
)
