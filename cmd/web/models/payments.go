package models

import "cchoice/internal/payments"

type AvailablePaymentMethod struct {
	ImageData string
	Value     payments.PaymentMethod
	Enabled   bool
}
