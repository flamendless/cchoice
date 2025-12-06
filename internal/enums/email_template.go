package enums

//go:generate go tool stringer -type=EmailTemplateName -trimprefix=EMAIL_TEMPLATE_

type EmailTemplateName int

const (
	EMAIL_TEMPLATE_UNDEFINED EmailTemplateName = iota
	EMAIL_TEMPLATE_ORDER_CONFIRMATION
	EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
)

func ParseEmailTemplateNameToEnum(e string) EmailTemplateName {
	switch e {
	case EMAIL_TEMPLATE_ORDER_CONFIRMATION.String():
		return EMAIL_TEMPLATE_ORDER_CONFIRMATION
	case EMAIL_TEMPLATE_PAYMENT_CONFIRMATION.String():
		return EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
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
	default:
		return EMAIL_TEMPLATE_UNDEFINED
	}
}
