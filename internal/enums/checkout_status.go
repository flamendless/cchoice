package enums

//go:generate go tool stringer -type=CheckoutStatus -trimprefix=CHECKOUT_STATUS_

type CheckoutStatus int

const (
	CHECKOUT_STATUS_UNDEFINED CheckoutStatus = iota
	CHECKOUT_STATUS_PENDING
	CHECKOUT_STATUS_COMPLETED
	CHECKOUT_STATUS_CANCELLED
)

func ParseCheckoutStatusToEnum(e string) CheckoutStatus {
	switch e {
	case CHECKOUT_STATUS_PENDING.String():
		return CHECKOUT_STATUS_PENDING
	case CHECKOUT_STATUS_COMPLETED.String():
		return CHECKOUT_STATUS_COMPLETED
	case CHECKOUT_STATUS_CANCELLED.String():
		return CHECKOUT_STATUS_CANCELLED
	default:
		return CHECKOUT_STATUS_UNDEFINED
	}
}
