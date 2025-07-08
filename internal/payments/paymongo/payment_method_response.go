package paymongo

import (
	"cchoice/internal/payments"
)

type GetAvailablePaymentMethodsResponse struct {
	Data []string
}

func (r GetAvailablePaymentMethodsResponse) ToPaymentMethods() []payments.PaymentMethod {
	res := make([]payments.PaymentMethod, 0, len(r.Data))
	for _, pm := range r.Data {
		res = append(res, payments.ParsePaymentMethodToEnum(pm))
	}
	return res
}

var _ payments.GetAvailablePaymentMethodsResponse = (*GetAvailablePaymentMethodsResponse)(nil)
