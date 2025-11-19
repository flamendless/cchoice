package cmd

import (
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/receipt"
	"cchoice/internal/receipt/scanner"
	"cchoice/internal/receipt/scanner/googlevision"
	"cchoice/internal/receipt/writer"
	csvwriter "cchoice/internal/receipt/writer/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scanImagePath  string
	scanOutputPath string
	scanOutputFmt  string
	scanCSV        bool
	scanJSON       bool
	ocrServiceName string
)

func init() {
	scanReceiptCmd.Flags().StringVarP(&ocrServiceName, "service", "s", "GOOGLEVISION", "OCR service to use")
	scanReceiptCmd.Flags().StringVarP(&scanImagePath, "image", "i", "", "Path to receipt image file (required)")
	scanReceiptCmd.Flags().StringVarP(&scanOutputPath, "output", "o", "", "Output file path (required for csv/json)")
	scanReceiptCmd.Flags().StringVarP(&scanOutputFmt, "format", "f", "", "Output format: csv, json, or both (comma-separated)")
	scanReceiptCmd.Flags().BoolVar(&scanCSV, "csv", false, "Save output as CSV")
	scanReceiptCmd.Flags().BoolVar(&scanJSON, "json", false, "Save output as JSON")

	if err := scanReceiptCmd.MarkFlagRequired("image"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(scanReceiptCmd)
}

var scanReceiptCmd = &cobra.Command{
	Use:   "scan_receipt",
	Short: "Scan receipt images",
	Long:  "Scan receipt images to extract data using OCR service",
	Run: func(cmd *cobra.Command, args []string) {
		var ocrService scanner.IReceiptScanner
		if ocrServiceName == receipt.RECEIPT_SCANNER_GOOGLEVISION.String() {
			ocrService = googlevision.MustInit()
			defer func(){
				if err := ocrService.Close(); err != nil {
					panic(err)
				}
			}()
		} else {
			panic("OCR service not currently supported")
		}

		if err := runScanReceipt(ocrService); err != nil {
			logs.Log().Error("Receipt scan failed", zap.Error(err))
			panic(err)
		}
	},
}

func runScanReceipt(ocrService scanner.IReceiptScanner) error {
	if scanImagePath == "" {
		return errors.Join(errs.ErrCmdInvalidFlag, errors.New("image path is required"))
	}

	if _, err := os.Stat(scanImagePath); os.IsNotExist(err) {
		return errors.Join(errs.ErrReceiptImageNotFound, fmt.Errorf("image file not found: %s", scanImagePath))
	}

	logs.Log().Info("Starting receipt scan", zap.String("image", scanImagePath))
	logs.Log().Info("Scanning receipt image")

	receiptData, err := ocrService.ScanReceipt(scanImagePath)
	if err != nil {
		return errors.Join(errs.ErrReceiptParsingFailed, err)
	}

	logs.Log().Info("Receipt scanned successfully")

	printWriter := writer.NewPrintWriter()
	if err := printWriter.Write(receiptData, ""); err != nil {
		logs.Log().Warn("Failed to print receipt data", zap.Error(err))
	}

	formats := parseOutputFormats()

	if len(formats) > 0 {
		if scanOutputPath == "" {
			return errors.Join(errs.ErrCmdInvalidFlag, errors.New("output path is required when saving to file"))
		}

		for _, format := range formats {
			if err := writeOutput(receiptData, format); err != nil {
				logs.Log().Error("Failed to write output", zap.String("format", format), zap.Error(err))
				return err
			}
		}
	}

	logs.Log().Info("Receipt scan completed successfully")
	return nil
}

func parseOutputFormats() []string {
	formats := []string{}

	if scanOutputFmt != "" {
		parts := strings.Split(scanOutputFmt, ",")
		for _, part := range parts {
			format := strings.TrimSpace(strings.ToLower(part))
			if format == "csv" || format == "json" {
				formats = append(formats, format)
			}
		}
	}

	if scanCSV && !contains(formats, "csv") {
		formats = append(formats, "csv")
	}
	if scanJSON && !contains(formats, "json") {
		formats = append(formats, "json")
	}

	return formats
}

func writeOutput(data *scanner.ReceiptData, format string) error {
	outputPath := scanOutputPath

	ext := filepath.Ext(outputPath)
	if ext == "" {
		outputPath = outputPath + "." + format
	} else if ext != "."+format {
		baseName := strings.TrimSuffix(outputPath, ext)
		outputPath = baseName + "." + format
	}

	switch format {
	case "csv":
		csvWriter := csvwriter.NewCSVWriter()
		return csvWriter.Write(data, outputPath)
	case "json":
		jsonWriter := writer.NewJSONWriter()
		return jsonWriter.Write(data, outputPath)
	default:
		return fmt.Errorf("%w: %s", errs.ErrReceiptInvalidFormat, format)
	}
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
