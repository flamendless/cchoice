package errs

import "errors"

var (
	ErrInvoice                   = errors.New("[INVOICE]: Error on invoice service")
	ErrInvoiceNotFound           = errors.New("[INVOICE]: Invoice not found")
	ErrInvoiceConfigRequired     = errors.New("[INVOICE]: Invoice configuration is required. Set it up first")
	ErrInvoiceNoLines            = errors.New("[INVOICE]: At least one line item is required")
	ErrInvoiceRecipientRequired  = errors.New("[INVOICE]: A recipient is required")
	ErrInvoiceRecipientNotFound  = errors.New("[INVOICE]: Recipient not found")
	ErrInvoiceRecipientNameReq   = errors.New("[INVOICE]: Recipient name is required")
	ErrInvoiceCreateFailed       = errors.New("[INVOICE]: Failed to create invoice")
	ErrInvoiceLogoInvalid        = errors.New("[INVOICE]: Logo must be a PNG or JPEG image")
	ErrInvoicePDFFailed          = errors.New("[INVOICE]: Failed to generate invoice PDF")
	ErrInvoiceEmailFailed        = errors.New("[INVOICE]: Failed to send invoice email")
	ErrInvoiceEmailNotConfigured = errors.New("[INVOICE]: Email service is not configured")
	ErrInvoiceRecipientNoEmail   = errors.New("[INVOICE]: Recipient has no email address")
)
