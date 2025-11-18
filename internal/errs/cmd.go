package errs

import "errors"

var (
	ErrCmd                     = errors.New("[CMD]: Error in command")
	ErrCmdRequired             = errors.New("[CMD]: Required flag")
	ErrCmdInvalidFlag          = errors.New("[CMD]: Invalid flag value")
	ErrCmdUnimplementedService = errors.New("[CMD]: Unimplemented service")
	ErrCmdInvalidService       = errors.New("[CMD]: Invalid service name")
	ErrCmdUndefinedService     = errors.New("[CMD]: Undefined service")
	ErrCmdInvalidFormat        = errors.New("[CMD]: Invalid format")
	ErrCmdMissingColumn        = errors.New("[CMD]: Missing required column")
	ErrCmdInvalidPrice         = errors.New("[CMD]: Invalid price data")
	ErrCmdNoCategory           = errors.New("[CMD]: Product has no category value")
	ErrCmdNoSubcategory        = errors.New("[CMD]: Product has no subcategory value")
)
