package enums

//go:generate go tool stringer -type=CustomerType -trimprefix=CUSTOMER_TYPE_

type CustomerType int

const (
	CUSTOMER_TYPE_UNDEFINED CustomerType = iota
	CUSTOMER_TYPE_CUSTOMER
	CUSTOMER_TYPE_COMPANY
)

func ParseCustomerTypeToEnum(e string) CustomerType {
	switch e {
	case CUSTOMER_TYPE_CUSTOMER.String():
		return CUSTOMER_TYPE_CUSTOMER
	case CUSTOMER_TYPE_COMPANY.String():
		return CUSTOMER_TYPE_COMPANY
	default:
		return CUSTOMER_TYPE_UNDEFINED
	}
}

func MustParseCustomerTypeToEnum(e string) CustomerType {
	switch e {
	case CUSTOMER_TYPE_CUSTOMER.String():
		return CUSTOMER_TYPE_CUSTOMER
	case CUSTOMER_TYPE_COMPANY.String():
		return CUSTOMER_TYPE_COMPANY
	default:
		panic("Invalid CustomerType. Got '" + e + "'")
	}
}

func (ct CustomerType) IsValid() bool {
	return ct != CUSTOMER_TYPE_UNDEFINED
}
