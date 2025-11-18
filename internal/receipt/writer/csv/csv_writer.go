package csv

import (
	"cchoice/internal/errs"
	"cchoice/internal/receipt/scanner"
	"cchoice/internal/receipt/writer"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CSVWriter struct{}

func NewCSVWriter() *CSVWriter {
	return &CSVWriter{}
}

func (w *CSVWriter) Write(data *scanner.ReceiptData, outputPath string) error {
	if data == nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("receipt data is nil"))
	}

	if outputPath == "" {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("output path is required for CSV format"))
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrFileCreate, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"RECEIPT INFORMATION"}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Field", "Value"}); err != nil {
		panic(err)
	}

	if err := writer.Write([]string{"Merchant Name", data.MerchantName}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Merchant Address", data.MerchantAddress}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Merchant Phone", data.MerchantPhone}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Merchant TIN", data.CustomerTIN}); err != nil {
		panic(err)
	}

	if err := writer.Write([]string{"Receipt Type", data.ReceiptType}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Receipt Number", data.ReceiptNumber}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Date", data.Date}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Time", data.Time}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Sold To", data.SoldTo}); err != nil {
		panic(err)
	}

	if err := writer.Write([]string{"Subtotal", data.Subtotal}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Total Sales VAT Inclusive", data.VATInclusive}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Less VAT", data.LessVAT}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Less Withholding Tax", data.LessWithholding}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Amount Net of VAT", data.AmountNetOfVAT}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Add VAT", data.AddVAT}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Tax", data.Tax}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Total", data.Total}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Payment Method", data.PaymentMethod}); err != nil {
		panic(err)
	}
	if err := writer.Write([]string{"Currency", data.Currency}); err != nil {
		panic(err)
	}

	if err := writer.Write([]string{}); err != nil {
		panic(err)
	}

	if len(data.Items) > 0 {
		if err := writer.Write([]string{"LINE ITEMS"}); err != nil {
			panic(err)
		}
		if err := writer.Write([]string{"#", "Item Name", "Quantity", "Price", "Subtotal"}); err != nil {
			panic(err)
		}

		for i, item := range data.Items {
			if err := writer.Write([]string{
				strconv.Itoa(i+1),
				item.Name,
				item.Quantity,
				item.Price,
				item.Subtotal,
			}); err != nil {
				panic(err)
			}
		}
	}

	if err := writer.Error(); err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrIOWrite, err)
	}

	fmt.Printf("Receipt data written to CSV file: %s\n", outputPath)
	return nil
}

func (w *CSVWriter) WriteFlattened(data *scanner.ReceiptData, outputPath string) error {
	if data == nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("receipt data is nil"))
	}

	if outputPath == "" {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("output path is required for CSV format"))
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrFileCreate, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Merchant Name",
		"Merchant Address",
		"Merchant Phone",
		"Merchant TIN",
		"Receipt Type",
		"Receipt Number",
		"Date",
		"Time",
		"Sold To",
		"Subtotal",
		"Total Sales VAT Inclusive",
		"Less VAT",
		"Less Withholding Tax",
		"Amount Net of VAT",
		"Add VAT",
		"Tax",
		"Total",
		"Payment Method",
		"Currency",
		"Items",
	}
	if err := writer.Write(headers); err != nil {
		panic(err)
	}

	itemsStr := make([]string, 0, len(data.Items))
	for _, item := range data.Items {
		itemStr := item.Name
		if item.Quantity != "" {
			itemStr += fmt.Sprintf(" (Qty: %s)", item.Quantity)
		}
		if item.Price != "" {
			itemStr += " @ " + item.Price
		}
		itemsStr = append(itemsStr, itemStr)
	}

	row := []string{
		data.MerchantName,
		data.MerchantAddress,
		data.MerchantPhone,
		data.CustomerTIN,
		data.ReceiptType,
		data.ReceiptNumber,
		data.Date,
		data.Time,
		data.SoldTo,
		data.Subtotal,
		data.VATInclusive,
		data.LessVAT,
		data.LessWithholding,
		data.AmountNetOfVAT,
		data.AddVAT,
		data.Tax,
		data.Total,
		data.PaymentMethod,
		data.Currency,
		strings.Join(itemsStr, "; "),
	}
	if err := writer.Write(row); err != nil {
		panic(err)
	}

	if err := writer.Error(); err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrIOWrite, err)
	}

	fmt.Printf("Receipt data written to CSV file (flattened): %s\n", outputPath)
	return nil
}

var _ writer.IReceiptWriter = (*CSVWriter)(nil)
