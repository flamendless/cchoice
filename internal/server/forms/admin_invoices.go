package forms

type AdminInvoicePath struct {
	ID string `param:"id" validate:"required"`
}

type AdminInvoiceRecipientPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminInvoiceRecipientForm struct {
	Name          string `form:"name" validate:"required"`
	Email         string `form:"email"`
	ContactNumber string `form:"contact_number"`
	Address       string `form:"address"`
	TIN           string `form:"tin"`
	Notes         string `form:"notes"`
}

type AdminInvoiceRecipientsQuery struct {
	Search string `form:"search"`
}

// AdminInvoiceLineInput is one line item in a create-invoice request.
type AdminInvoiceLineInput struct {
	ProductID   string `json:"product_id"`
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Quantity    int64  `json:"quantity"`
}

// AdminInvoiceNewRecipientInput carries an inline new recipient to persist.
type AdminInvoiceNewRecipientInput struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Address       string `json:"address"`
	TIN           string `json:"tin"`
	Notes         string `json:"notes"`
}

// AdminInvoiceCreateForm is the JSON payload sent by the generate-invoice modal.
type AdminInvoiceCreateForm struct {
	RecipientID  string                         `json:"recipient_id"`
	NewRecipient *AdminInvoiceNewRecipientInput `json:"new_recipient"`
	Notes        string                         `json:"notes"`
	DueDate      string                         `json:"due_date"`
	Action       string                         `json:"action"`
	Lines        []AdminInvoiceLineInput        `json:"lines"`
}
