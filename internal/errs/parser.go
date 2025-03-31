package errs

import (
	stdErrors "errors"
	"fmt"
)

type ParserErrorCode string

const (
	ProductDiscontinued ParserErrorCode = "PRODUCT_DISCONTINUED"
	BlankProductName    ParserErrorCode = "BLANK_PRODUCT_NAME"
	CantCovert          ParserErrorCode = "CANT_CONVERT"
)

type parserError struct {
	error
	errorCode ParserErrorCode
}

func (pe parserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.errorCode, pe.error.Error())
}

func ParserErrorCodeUnwrap(err error) error {
	e, ok := err.(parserError)
	if ok {
		return stdErrors.Unwrap(e.error)
	}
	return stdErrors.Unwrap(err)
}

func ParserErrorCodeCode(err error) ParserErrorCode {
	if err == nil {
		return ""
	}

	e, ok := err.(parserError)
	if ok {
		return e.errorCode
	}

	return ""
}

func NewParserError(errorCode ParserErrorCode, format string, args ...any) error {
	return parserError{
		error:     fmt.Errorf(format, args...),
		errorCode: errorCode,
	}
}

func WrapIntoParserError(err error, errorCode ParserErrorCode, msg string) error {
	return parserError{
		error:     fmt.Errorf("%s: [%w]", msg, err),
		errorCode: errorCode,
	}
}
