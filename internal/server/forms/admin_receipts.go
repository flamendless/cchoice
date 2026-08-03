package forms

type AdminReceiptPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminReceiptsTableQuery struct {
	Page int `form:"page"`
}

type AdminReceiptGenerateQuery struct {
	InvoiceID string `form:"invoice_id"`
}

type AdminDeliveryReceiptLineInput struct {
	Quantity    string `json:"quantity"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}

type AdminDeliveryReceiptCreateForm struct {
	InvoiceID      string                          `json:"invoice_id"`
	DeliveredTo    string                          `json:"delivered_to"`
	RecipientEmail string                          `json:"recipient_email"`
	TIN            string                          `json:"tin"`
	Address        string                          `json:"address"`
	ReceiptDate    string                          `json:"receipt_date"`
	Terms          string                          `json:"terms"`
	PONumber       string                          `json:"po_number"`
	Action         string                          `json:"action"`
	Lines          []AdminDeliveryReceiptLineInput `json:"lines"`
}

type AdminCollectionReceiptSettlementInput struct {
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
	Amount        string `json:"amount"`
}

type AdminCollectionReceiptCreateForm struct {
	InvoiceID      string                                  `json:"invoice_id"`
	ReceivedFrom   string                                  `json:"received_from"`
	RecipientEmail string                                `json:"recipient_email"`
	TIN            string                                `json:"tin"`
	Address        string                                `json:"address"`
	ReceiptDate    string                                `json:"receipt_date"`
	Amount         string                                `json:"amount"`
	PaymentFor     string                                `json:"payment_for"`
	PaymentForm    string                                `json:"payment_form"`
	SCCitizenTIN   string                                `json:"sc_citizen_tin"`
	OSCAPWDIDNo    string                                `json:"osca_pwd_id_no"`
	Action         string                                `json:"action"`
	Settlements    []AdminCollectionReceiptSettlementInput `json:"settlements"`
}
