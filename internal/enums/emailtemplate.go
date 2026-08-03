package enums

//go:generate go tool stringer -type=EmailTemplateName -trimprefix=EMAIL_TEMPLATE_

type EmailTemplateName int

const (
	EMAIL_TEMPLATE_UNDEFINED EmailTemplateName = iota
	EMAIL_TEMPLATE_ORDER_CONFIRMATION
	EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
	EMAIL_TEMPLATE_CUSTOMER_VERIFICATION
	EMAIL_TEMPLATE_PASSWORD_RESET
	EMAIL_TEMPLATE_MEMO_NOTIFICATION
	EMAIL_TEMPLATE_ORDER_STATUS_UPDATE
	EMAIL_TEMPLATE_INVOICE
	EMAIL_TEMPLATE_DELIVERY_RECEIPT
	EMAIL_TEMPLATE_COLLECTION_RECEIPT
)

func ParseEmailTemplateNameToEnum(e string) EmailTemplateName {
	switch e {
	case EMAIL_TEMPLATE_ORDER_CONFIRMATION.String():
		return EMAIL_TEMPLATE_ORDER_CONFIRMATION
	case EMAIL_TEMPLATE_PAYMENT_CONFIRMATION.String():
		return EMAIL_TEMPLATE_PAYMENT_CONFIRMATION
	case EMAIL_TEMPLATE_PASSWORD_RESET.String():
		return EMAIL_TEMPLATE_PASSWORD_RESET
	case EMAIL_TEMPLATE_MEMO_NOTIFICATION.String():
		return EMAIL_TEMPLATE_MEMO_NOTIFICATION
	case EMAIL_TEMPLATE_ORDER_STATUS_UPDATE.String():
		return EMAIL_TEMPLATE_ORDER_STATUS_UPDATE
	case EMAIL_TEMPLATE_INVOICE.String():
		return EMAIL_TEMPLATE_INVOICE
	case EMAIL_TEMPLATE_DELIVERY_RECEIPT.String():
		return EMAIL_TEMPLATE_DELIVERY_RECEIPT
	case EMAIL_TEMPLATE_COLLECTION_RECEIPT.String():
		return EMAIL_TEMPLATE_COLLECTION_RECEIPT
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
	case EMAIL_TEMPLATE_MEMO_NOTIFICATION:
		return "memo_notification.html"
	case EMAIL_TEMPLATE_ORDER_STATUS_UPDATE:
		return "order_status_update.html"
	case EMAIL_TEMPLATE_INVOICE:
		return "invoice_email.html"
	case EMAIL_TEMPLATE_DELIVERY_RECEIPT:
		return "delivery_receipt_email.html"
	case EMAIL_TEMPLATE_COLLECTION_RECEIPT:
		return "collection_receipt_email.html"
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
	case EMAIL_TEMPLATE_MEMO_NOTIFICATION:
		return "memo_notification"
	case EMAIL_TEMPLATE_ORDER_STATUS_UPDATE:
		return "order_status_update"
	case EMAIL_TEMPLATE_INVOICE:
		return "invoice"
	case EMAIL_TEMPLATE_DELIVERY_RECEIPT:
		return "delivery_receipt"
	case EMAIL_TEMPLATE_COLLECTION_RECEIPT:
		return "collection_receipt"
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
	case "memo_notification":
		return EMAIL_TEMPLATE_MEMO_NOTIFICATION
	case "order_status_update":
		return EMAIL_TEMPLATE_ORDER_STATUS_UPDATE
	case "invoice":
		return EMAIL_TEMPLATE_INVOICE
	case "delivery_receipt":
		return EMAIL_TEMPLATE_DELIVERY_RECEIPT
	case "collection_receipt":
		return EMAIL_TEMPLATE_COLLECTION_RECEIPT
	default:
		return EMAIL_TEMPLATE_UNDEFINED
	}
}
