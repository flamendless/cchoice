package forms

import (
	"strings"

	"cchoice/internal/constants"
)

type AdminInvoicePath struct {
	ID string `param:"id" validate:"required"`
}

type AdminInvoiceRecipientPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminInvoiceRecipientForm struct {
	Name           string `form:"name" validate:"required"`
	Email          string `form:"email"`
	ContactNumber  string `form:"contact_number"`
	Address        string `form:"address"`
	TIN            string `form:"tin"`
	RegisteredName string `form:"registered_name"`
	Notes          string `form:"notes"`
}

type AdminInvoiceRecipientsQuery struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
}

type AdminInvoicesTableQuery struct {
	Page int `form:"page"`
}

type AdminInvoiceLineInput struct {
	ProductID   string `json:"product_id"`
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Quantity    int64  `json:"quantity"`
	TaxType     string `json:"tax_type"`
}

type AdminInvoiceNewRecipientInput struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	ContactNumber  string `json:"contact_number"`
	Address        string `json:"address"`
	TIN            string `json:"tin"`
	RegisteredName string `json:"registered_name"`
	Notes          string `json:"notes"`
}

type AdminInvoiceCreateForm struct {
	RecipientID             string                         `json:"recipient_id"`
	NewRecipient            *AdminInvoiceNewRecipientInput `json:"new_recipient"`
	TransactionType         string                         `json:"transaction_type"`
	RecipientRegisteredName string                         `json:"recipient_registered_name"`
	Notes                   string                         `json:"notes"`
	DueDate                 string                         `json:"due_date"`
	WithholdingTax          string                         `json:"withholding_tax"`
	SCPWDDiscount           string                         `json:"sc_pwd_discount"`
	AddVAT                  string                         `json:"add_vat"`
	Action                  string                         `json:"action"`
	Lines                   []AdminInvoiceLineInput        `json:"lines"`
}

func (f *AdminInvoiceRecipientForm) Normalize() {
	f.ContactNumber = normalizePHMobile(f.ContactNumber)
}

func (f *AdminInvoiceNewRecipientInput) Normalize() {
	f.ContactNumber = normalizePHMobile(f.ContactNumber)
}

func normalizePHMobile(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, constants.PHMobilePrefix) {
		s = constants.PHMobilePrefix + strings.TrimPrefix(s, "0")
	}
	return s
}
