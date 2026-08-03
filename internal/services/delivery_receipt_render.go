package services

import (
	"cchoice/internal/constants"
)

func (s *DeliveryReceiptService) BuildRenderData(config InvoiceConfig, receipt DeliveryReceipt, lines []DeliveryReceiptLine) DeliveryReceiptRenderData {
	return DeliveryReceiptRenderData{Config: config, Receipt: receipt, Lines: lines}
}

func (s *DeliveryReceiptService) BuildEmailTemplateData(config InvoiceConfig, receipt DeliveryReceipt, lines []DeliveryReceiptLine) map[string]any {
	lineItems := make([]map[string]any, 0, len(lines))
	for _, l := range lines {
		lineItems = append(lineItems, map[string]any{
			"Quantity":    l.Quantity,
			"Unit":        l.Unit,
			"Description": l.Description,
		})
	}
	businessName := cmpOr(config.BusinessName, "C-Choice")
	return map[string]any{
		"LogoURL":        cmpOr(config.LogoURL, constants.PathEmailLogoCDN),
		"BusinessName":   businessName,
		"ReceiptNumber":  receipt.ReceiptNumber,
		"DeliveredTo":    receipt.DeliveredTo,
		"RecipientEmail": receipt.RecipientEmail,
		"ReceiptDate":    receipt.ReceiptDate,
		"Terms":          receipt.Terms,
		"PONumber":       receipt.PONumber,
		"InvoiceNumber":  receipt.InvoiceNumber,
		"LineItems":      lineItems,
		"MobileNo":       config.ContactNumber,
		"EMail":          config.Email,
	}
}
