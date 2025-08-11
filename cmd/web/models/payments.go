package models

import "cchoice/internal/payments"

type AvailablePaymentMethod struct {
	Value     payments.PaymentMethod
	ImageData string
	Enabled   bool
}
