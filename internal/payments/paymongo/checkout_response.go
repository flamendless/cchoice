package paymongo

import (
	"cchoice/internal/database/queries"
	"cchoice/internal/payments"
	"strings"
	"time"
)

type CreateCheckoutSessionResponse struct {
	Data CheckoutSession `json:"data"`
}

func (r *CreateCheckoutSessionResponse) ToCheckout(
	pg payments.PaymentGateway,
) *queries.CreateCheckoutParams {
	var paidAt time.Time
	if len(r.Data.Attributes.PaymentIntent.Attributes.Payments) > 0 {
		paidAt = time.Unix(int64(r.Data.Attributes.PaymentIntent.Attributes.Payments[0].Attributes.PaidAt), 0)
	}

	return &queries.CreateCheckoutParams{
		ID:                     r.Data.ID,
		Gateway:                pg.GatewayName(),
		Status:                 r.Data.Attributes.Status,
		Description:            r.Data.Attributes.Description,
		TotalAmount:            int64(r.Data.Attributes.PaymentIntent.Attributes.Amount),
		CheckoutUrl:            r.Data.Attributes.CheckoutURL,
		ClientKey:              r.Data.Attributes.ClientKey,
		ReferenceNumber:        r.Data.Attributes.ReferenceNumber,
		PaymentStatus:          r.Data.Attributes.PaymentIntent.Attributes.Status,
		PaymentMethodType:      strings.Join(r.Data.Attributes.PaymentMethodTypes, ","),
		PaidAt:                 paidAt,
		MetadataRemarks:        r.Data.Attributes.Metadata.Remarks,
		MetadataNotes:          r.Data.Attributes.Metadata.Notes,
		MetadataCustomerNumber: r.Data.Attributes.Metadata.CustomerNumber,
	}
}

func (r *CreateCheckoutSessionResponse) ToLineItems() []*queries.CreateCheckoutLineItemParams {
	res := make([]*queries.CreateCheckoutLineItemParams, 0, len(r.Data.Attributes.LineItems))
	for _, lineItem := range r.Data.Attributes.LineItems {
		res = append(res, &queries.CreateCheckoutLineItemParams{
			CheckoutID:  r.Data.ID,
			Amount:      int64(lineItem.Amount),
			Currency:    lineItem.Currency,
			Description: lineItem.Description,
			Name:        lineItem.Name,
			Quantity:    int64(lineItem.Quantity),
		})
	}
	return res
}

var _ payments.CreateCheckoutSessionResponse = (*CreateCheckoutSessionResponse)(nil)
