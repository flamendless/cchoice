package errs

import "errors"

var (
	ErrImportPreviewExpired = errors.New("[IMPORT]: Import preview expired, please upload the file again")
	ErrImportNoRowsSelected = errors.New("[IMPORT]: Select at least one row to apply")
)
