package models

type CartShippingPrefill struct {
	Email        string
	FullName     string
	MobileNo     string
	AddressLine1 string
	AddressLine2 string
	Postal       string
	Province     string
	City         string
	Barangay     string
}

func (p CartShippingPrefill) HasData() bool {
	return p.Email != "" ||
		p.FullName != "" ||
		p.MobileNo != "" ||
		p.AddressLine1 != "" ||
		p.AddressLine2 != "" ||
		p.Postal != "" ||
		p.Province != "" ||
		p.City != "" ||
		p.Barangay != ""
}
