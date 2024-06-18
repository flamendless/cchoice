package grpc

import (
	stdErrors "errors"
	"fmt"
)

type ErrorCode string

const (
	IDNotFound            ErrorCode = "ID_NOT_FOUND"
	UnimplementedFunction ErrorCode = "UNIMPLEMENTED_FUNC"
)

type grpcError struct {
	error
	errorCode ErrorCode
}

func (pe grpcError) Error() string {
	return fmt.Sprintf("%s: %s", pe.errorCode, pe.error.Error())
}

func Unwrap(err error) error {
	e, ok := err.(grpcError)
	if ok {
		return stdErrors.Unwrap(e.error)
	}
	return stdErrors.Unwrap(err)
}

func Code(err error) ErrorCode {
	if err == nil {
		return ""
	}

	e, ok := err.(grpcError)
	if ok {
		return e.errorCode
	}

	return ""
}

func NewGRPCError(errorCode ErrorCode, format string, args ...interface{}) error {
	return grpcError{
		error:     fmt.Errorf(format, args...),
		errorCode: errorCode,
	}
}

func WrapIntoGRPCError(err error, errorCode ErrorCode, msg string) error {
	return grpcError{
		error:     fmt.Errorf("%s: [%w]", msg, err),
		errorCode: errorCode,
	}
}
