package errs

import (
	stdErrors "errors"
	"fmt"
)

type GRPCErrorCode string

const (
	IDNotFound            GRPCErrorCode = "ID_NOT_FOUND"
	UnimplementedFunction GRPCErrorCode = "UNIMPLEMENTED_FUNC"
)

type grpcError struct {
	error
	errorCode GRPCErrorCode
}

func (pe grpcError) Error() string {
	return fmt.Sprintf("%s: %s", pe.errorCode, pe.error.Error())
}

func GRPCErrorCodeUnwrap(err error) error {
	e, ok := err.(grpcError)
	if ok {
		return stdErrors.Unwrap(e.error)
	}
	return stdErrors.Unwrap(err)
}

func GRPCErrorCodeCode(err error) GRPCErrorCode {
	if err == nil {
		return ""
	}

	e, ok := err.(grpcError)
	if ok {
		return e.errorCode
	}

	return ""
}

func NewGRPCError(errorCode GRPCErrorCode, format string, args ...interface{}) error {
	return grpcError{
		error:     fmt.Errorf(format, args...),
		errorCode: errorCode,
	}
}

func WrapIntoGRPCError(err error, errorCode GRPCErrorCode, msg string) error {
	return grpcError{
		error:     fmt.Errorf("%s: [%w]", msg, err),
		errorCode: errorCode,
	}
}
