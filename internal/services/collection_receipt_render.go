package services

import (
	"cchoice/internal/constants"
	"cchoice/internal/utils"
)

type CollectionReceiptRenderData struct {
	Config       InvoiceConfig
	Receipt      CollectionReceipt
	Settlements  []CollectionReceiptSettlement
}

func (s *CollectionReceiptService) BuildRenderData(config InvoiceConfig, receipt CollectionReceipt, settlements []CollectionReceiptSettlement) CollectionReceiptRenderData {
	return CollectionReceiptRenderData{Config: config, Receipt: receipt, Settlements: settlements}
}

func (s *CollectionReceiptService) BuildEmailTemplateData(config InvoiceConfig, receipt CollectionReceipt, settlements []CollectionReceiptSettlement) map[string]any {
	currency := cmpOr(config.Currency, constants.PHP)
	settlementRows := make([]map[string]any, 0, len(settlements))
	for _, st := range settlements {
		settlementRows = append(settlementRows, map[string]any{
			"InvoiceNumber": st.InvoiceNumber,
			"Amount":        st.Amount,
		})
	}
	businessName := cmpOr(config.BusinessName, "C-Choice")
	return map[string]any{
		"LogoURL":        cmpOr(config.LogoURL, constants.PathEmailLogoCDN),
		"BusinessName":   businessName,
		"ReceiptNumber":  receipt.ReceiptNumber,
		"ReceivedFrom":   receipt.ReceivedFrom,
		"ReceiptDate":    receipt.ReceiptDate,
		"Amount":         utils.NewMoney(receipt.AmountRaw, currency).Display(),
		"AmountInWords":  receipt.AmountInWords,
		"PaymentFor":     receipt.PaymentFor,
		"PaymentForm":    receipt.PaymentForm.String(),
		"Settlements":    settlementRows,
		"MobileNo":       config.ContactNumber,
		"EMail":          config.Email,
	}
}
