package cart

import "cchoice/internal/encode"

type CartCheckout struct {
	AddressLine1  string   `json:"address_line1"`
	AddressLine2  string   `json:"address_line2"`
	Province      string   `json:"province"`
	Postal        string   `json:"postal"`
	City          string   `json:"city"`
	PaymentMethod string   `json:"checked_payment_method"`
	Email         string   `json:"email"`
	MobileNo      string   `json:"mobile_no"`
	Barangay      string   `json:"barangay"`
	CheckoutIDs   []string `json:"checked_item"`
}

func (c *CartCheckout) ToDBCheckoutIDs(encode encode.IEncode) []int64 {
	dbCheckoutLineIDs := make([]int64, 0, len(c.CheckoutIDs))
	for _, checkoutLineID := range c.CheckoutIDs {
		dbCheckoutLineIDs = append(
			dbCheckoutLineIDs,
			encode.Decode(checkoutLineID),
		)
	}
	return dbCheckoutLineIDs
}
