package models

import "cchoice/internal/payments"

type AvailablePaymentMethod struct {
	ImageURL string
	Value    payments.PaymentMethod
	Enabled  bool
}
