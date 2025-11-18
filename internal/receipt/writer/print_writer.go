package writer

import (
	"cchoice/internal/errs"
	"cchoice/internal/receipt/scanner"
	"errors"
	"fmt"
)

type PrintWriter struct{}

func NewPrintWriter() *PrintWriter {
	return &PrintWriter{}
}

func (w *PrintWriter) Write(data *scanner.ReceiptData, outputPath string) error {
	if data == nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("receipt data is nil"))
	}

	fmt.Println("========================================")
	fmt.Println("RECEIPT SCAN RESULTS")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("MERCHANT INFORMATION")
	fmt.Println("--------------------")
	fmt.Printf("Name:     %s\n", valueOrEmpty(data.MerchantName))
	fmt.Printf("Address:  %s\n", valueOrEmpty(data.MerchantAddress))
	fmt.Printf("Phone:    %s\n", valueOrEmpty(data.MerchantPhone))
	fmt.Printf("TIN:      %s\n", valueOrEmpty(data.MerchantTIN))
	fmt.Printf("Prop:     %s\n", valueOrEmpty(data.MerchantProp))
	fmt.Println()

	fmt.Println("TRANSACTION DETAILS")
	fmt.Println("-------------------")
	fmt.Printf("Type:               %s\n", valueOrEmpty(data.ReceiptType))
	fmt.Printf("Receipt #:          %s\n", valueOrEmpty(data.ReceiptNumber))
	fmt.Printf("Date:               %s\n", valueOrEmpty(data.Date))
	fmt.Printf("Time:               %s\n", valueOrEmpty(data.Time))
	fmt.Printf("Sold To:            %s\n", valueOrEmpty(data.SoldTo))
	fmt.Printf("Customer TIN:       %s\n", valueOrEmpty(data.CustomerTIN))
	fmt.Printf("Customer Address:   %s\n", valueOrEmpty(data.CustomerAddress))
	fmt.Println()

	if len(data.Items) > 0 {
		fmt.Println("LINE ITEMS")
		fmt.Println("----------")
		for i, item := range data.Items {
			fmt.Printf("%d. %s\n", i+1, item.Name)
			fmt.Printf("   Quantity: %s\n", item.Quantity)
			fmt.Printf("   Price:    %s\n", item.Price)
			fmt.Printf("   Subtotal: %s\n", item.Subtotal)
		}
		fmt.Println()
	}

	fmt.Println("PAYMENT INFORMATION")
	fmt.Println("-------------------")
	fmt.Printf("Subtotal:                %s\n", valueOrEmpty(data.Subtotal))
	fmt.Printf("Total Sales VAT Incl:    %s\n", valueOrEmpty(data.VATInclusive))
	fmt.Printf("Less VAT:                %s\n", valueOrEmpty(data.LessVAT))
	fmt.Printf("Less Withholding Tax:    %s\n", valueOrEmpty(data.LessWithholding))
	fmt.Printf("Amount Net of VAT:       %s\n", valueOrEmpty(data.AmountNetOfVAT))
	fmt.Printf("Add VAT:                 %s\n", valueOrEmpty(data.AddVAT))
	fmt.Printf("Tax:                     %s\n", valueOrEmpty(data.Tax))
	fmt.Printf("Total:                   %s\n", valueOrEmpty(data.Total))
	fmt.Printf("Method:                  %s\n", valueOrEmpty(data.PaymentMethod))
	fmt.Printf("Currency:                %s\n", valueOrEmpty(data.Currency))
	fmt.Println()

	fmt.Println("========================================")

	return nil
}

var _ IReceiptWriter = (*PrintWriter)(nil)
