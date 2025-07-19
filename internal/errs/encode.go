package errs

import "errors"

var (
	ERR_DECODE = errors.New("[ENCODE]: Error in decoding")
	ERR_ENCODE = errors.New("[ENCODE]: Error in encoding")
)
