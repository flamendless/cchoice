package services

import (
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/utils"
)

type InvoiceRenderData struct {
	Config  InvoiceConfig
	Invoice Invoice
	Lines   []InvoiceLine
}

func (s *InvoiceService) BuildRenderData(config InvoiceConfig, invoice Invoice, lines []InvoiceLine) InvoiceRenderData {
	return InvoiceRenderData{Config: config, Invoice: invoice, Lines: lines}
}

func (s *InvoiceService) BuildEmailTemplateData(config InvoiceConfig, invoice Invoice, lines []InvoiceLine) map[string]any {
	currency := invoice.Currency
	if currency == "" {
		currency = constants.PHP
	}
	lineItems := make([]map[string]any, 0, len(lines))
	for _, l := range lines {
		lineItems = append(lineItems, map[string]any{
			"Description": l.Description,
			"Quantity":    l.Quantity,
			"UnitPrice":   l.UnitPrice,
			"LineTotal":   l.LineTotal,
		})
	}
	businessName := cmpOr(config.BusinessName, "C-Choice")
	return map[string]any{
		"LogoURL":       cmpOr(config.LogoURL, constants.PathEmailLogoCDN),
		"BusinessName":  businessName,
		"InvoiceNumber": invoice.InvoiceNumber,
		"RecipientName": invoice.RecipientName,
		"IssueDate":     invoice.IssueDate,
		"DueDate":       invoice.DueDate,
		"LineItems":     lineItems,
		"Subtotal":      utils.NewMoney(invoice.TotalSalesRaw, currency).Display(),
		"VATAmount":     invoice.VATAmount,
		"VATPercentage": invoice.VATPercentage,
		"Total":         utils.NewMoney(invoice.TotalSalesVATInclusiveRaw, currency).Display(),
		"Notes":         invoice.Notes,
		"MobileNo":      config.ContactNumber,
		"EMail":         config.Email,
	}
}

func defaultTaxType(t string) enums.InvoiceLineTaxType {
	if tt := enums.ParseInvoiceLineTaxTypeToEnum(t); tt != enums.INVOICE_LINE_TAX_UNDEFINED {
		return tt
	}
	return enums.INVOICE_LINE_TAX_VATABLE
}

func defaultTransactionType(t string) enums.InvoiceTransactionType {
	if tt := enums.ParseInvoiceTransactionTypeToEnum(t); tt != enums.INVOICE_TRANSACTION_UNDEFINED {
		return tt
	}
	return enums.INVOICE_TRANSACTION_CASH_SALES
}
