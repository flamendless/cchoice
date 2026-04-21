package errs

import "errors"

var (
	ErrProductInventory         = errors.New("[PRODUCT INVENTORY]: Error on product inventory service")
	ErrProductInventoryNotFound = errors.New("[PRODUCT INVENTORY]: Product inventory not found")
)
