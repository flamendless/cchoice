package forms

import "cchoice/internal/errs"

type ShippingAddressQuery struct {
	Data     string `form:"data" validate:"required,oneof=provinces cities barangays"`
	Province string `form:"province"`
	City     string `form:"city"`
}

func (q ShippingAddressQuery) Validate() error {
	switch q.Data {
	case "cities":
		if q.Province == "" {
			return errs.ErrInvalidParams
		}
	case "barangays":
		if q.City == "" {
			return errs.ErrInvalidParams
		}
	}
	return nil
}

const ncrProvince = "National Capital Region (NCR)"

type ShippingQuotationForm struct {
	AddressLine1 string `form:"address_line1"`
	AddressLine2 string `form:"address_line2"`
	City         string `form:"city"`
	Province     string `form:"province"`
	Barangay     string `form:"barangay"`
	Postal       string `form:"postal"`
}

func (f ShippingQuotationForm) Validate() error {
	if f.Province == ncrProvince {
		return nil
	}
	if f.Province == "" || f.City == "" || f.Barangay == "" {
		return errs.ErrInvalidParams
	}
	return nil
}
