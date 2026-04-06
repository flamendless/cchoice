package enums

//go:generate go tool stringer -type=CustomerStatus -trimprefix=CUSTOMER_STATUS_

type CustomerStatus int

const (
	CUSTOMER_STATUS_UNVERIFIED CustomerStatus = iota
	CUSTOMER_STATUS_VERIFIED
)

func ParseCustomerStatusToEnum(e string) CustomerStatus {
	switch e {
	case CUSTOMER_STATUS_VERIFIED.String():
		return CUSTOMER_STATUS_VERIFIED
	case CUSTOMER_STATUS_UNVERIFIED.String():
		return CUSTOMER_STATUS_UNVERIFIED
	default:
		return CUSTOMER_STATUS_UNVERIFIED
	}
}

func MustParseCustomerStatusToEnum(e string) CustomerStatus {
	switch e {
	case CUSTOMER_STATUS_VERIFIED.String():
		return CUSTOMER_STATUS_VERIFIED
	case CUSTOMER_STATUS_UNVERIFIED.String():
		return CUSTOMER_STATUS_UNVERIFIED
	default:
		panic("Invalid CustomerStatus. Got '" + e + "'")
	}
}

func (cs CustomerStatus) IsValid() bool {
	return cs == CUSTOMER_STATUS_VERIFIED || cs == CUSTOMER_STATUS_UNVERIFIED
}
