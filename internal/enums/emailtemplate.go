package enums

//go:generate go tool stringer -type=EmailTemplateName -trimprefix=EMAIL_TEMPLATE_

type EmailTemplateName int

const (
	EMAIL_TEMPLATE_UNDEFINED EmailTemplateName = iota
	EMAIL_TEMPLATE_ORDER_CONFIRMATION
	EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
	EMAIL_TEMPLATE_CUSTOMER_VERIFICATION
	EMAIL_TEMPLATE_PASSWORD_RESET
)

func ParseEmailTemplateNameToEnum(e string) EmailTemplateName {
	switch e {
	case EMAIL_TEMPLATE_ORDER_CONFIRMATION.String():
		return EMAIL_TEMPLATE_ORDER_CONFIRMATION
	case EMAIL_TEMPLATE_PAYMENT_CONFIRMATION.String():
		return EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
	case EMAIL_TEMPLATE_CUSTOMER_VERIFICATION.String():
		return EMAIL_TEMPLATE_CUSTOMER_VERIFICATION
	default:
		return EMAIL_TEMPLATE_UNDEFINED
	}
}

func (e EmailTemplateName) FileName() string {
	switch e {
	case EMAIL_TEMPLATE_ORDER_CONFIRMATION:
		return "order_confirmation.html"
	case EMAIL_TEMPLATE_PAYMENT_CONFIRMATION:
		return "payment_confirmation.html"
	case EMAIL_TEMPLATE_CUSTOMER_VERIFICATION:
		return "customer_verification.html"
	case EMAIL_TEMPLATE_PASSWORD_RESET:
		return "password_reset.html"
	default:
		return ""
	}
}

func (e EmailTemplateName) DBValue() string {
	switch e {
	case EMAIL_TEMPLATE_ORDER_CONFIRMATION:
		return "order_confirmation"
	case EMAIL_TEMPLATE_PAYMENT_CONFIRMATION:
		return "payment_confirmation"
	case EMAIL_TEMPLATE_CUSTOMER_VERIFICATION:
		return "customer_verification"
	case EMAIL_TEMPLATE_PASSWORD_RESET:
		return "password_reset"
	default:
		return ""
	}
}

func ParseEmailTemplateNameFromDB(s string) EmailTemplateName {
	switch s {
	case "order_confirmation":
		return EMAIL_TEMPLATE_ORDER_CONFIRMATION
	case "payment_confirmation":
		return EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
	case "customer_verification":
		return EMAIL_TEMPLATE_CUSTOMER_VERIFICATION
	case "password_reset":
		return EMAIL_TEMPLATE_PASSWORD_RESET
	default:
		return EMAIL_TEMPLATE_UNDEFINED
	}
}
