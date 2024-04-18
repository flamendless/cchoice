package parser

import (
	stdErrors "errors"
	"fmt"
)

type ErrorCode string

const (
	BlankField ErrorCode = "BLANK_FIELD"
)

type parserError struct {
	error
	errorCode ErrorCode
}

func (pe parserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.errorCode, pe.error.Error())
}

func Unwrap(err error) error {
	e, ok := err.(parserError)
	if ok {
		return stdErrors.Unwrap(e.error)
	}
	return stdErrors.Unwrap(err)
}

func Code(err error) ErrorCode {
	if err == nil {
		return ""
	}

	e, ok := err.(parserError)
	if ok {
		return e.errorCode
	}

	return ""
}

func NewParserError(errorCode ErrorCode, format string, args ...interface{}) error {
	return parserError{
		error:     fmt.Errorf(format, args...),
		errorCode: errorCode,
	}
}

func WrapIntoParserError(err error, errorCode ErrorCode, msg string) error {
	return parserError{
		error:     fmt.Errorf("%s: [%w]", msg, err),
		errorCode: errorCode,
	}
}
