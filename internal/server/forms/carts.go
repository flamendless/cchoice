package forms

import "cchoice/internal/errs"

type CartAddProductForm struct {
	ProductID string `form:"product_id" validate:"required"`
	Quantity  int64  `form:"quantity"`
}

type CartCheckoutLinePath struct {
	CheckoutLineID string `param:"checkoutline_id" validate:"required"`
}

type CartSummaryQuery struct {
	Data string `form:"data" validate:"required,oneof=summary_total"`
}

type CartUpdateQtyQuery struct {
	Dec string `form:"dec"`
	Inc string `form:"inc"`
}

func (q CartUpdateQtyQuery) Validate() error {
	if q.Dec == "1" || q.Inc == "1" {
		return nil
	}
	return errs.ErrInvalidParams
}
