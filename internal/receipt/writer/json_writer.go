package writer

import (
	"cchoice/internal/errs"
	"cchoice/internal/receipt/scanner"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type JSONWriter struct{}

func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

func (w *JSONWriter) Write(data *scanner.ReceiptData, outputPath string) error {
	if data == nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("receipt data is nil"))
	}

	if outputPath == "" {
		return errors.Join(errs.ErrReceiptWriteFailed, errors.New("output path is required for JSON format"))
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrJSONMarshal, err)
	}

	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		return errors.Join(errs.ErrReceiptWriteFailed, errs.ErrFileWrite, err)
	}

	fmt.Printf("Receipt data written to JSON file: %s\n", outputPath)
	return nil
}

var _ IReceiptWriter = (*JSONWriter)(nil)
