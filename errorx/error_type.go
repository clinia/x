package errorx

import (
	"fmt"
)

type ErrorType string

const (
	// The Invalid type should not be used, only useful to assert whether or not an error is an MdmError during cast
	ErrorTypeUnspecified   = ErrorType("")
	ErrorTypeInternal      = ErrorType("INTERNAL")
	ErrorTypeOutOfRange    = ErrorType("OUT_OF_RANGE")
	ErrorTypeUnsupported   = ErrorType("UNSUPPORTED") // Bad request
	ErrorTypeNotFound      = ErrorType("NOT_FOUND")
	ErrorTypeAlreadyExists = ErrorType("ALREADY_EXISTS") // Conflict
	// Map to 422
	ErrorTypeInvalidFormat = ErrorType("INVALID_FORMAT")
	// Map to 400
	ErrorTypeFailedPrecondition = ErrorType("FAILED_PRECONDITION")
	ErrorTypeNotImplemented     = ErrorType("NOT_IMPLEMENTED")
)

func ParseErrorType(s string) (ErrorType, error) {
	e := ErrorType(s)
	if err := e.Validate(); err != nil {
		return ErrorTypeUnspecified, err
	}

	return e, nil
}

func (e ErrorType) String() string {
	return string(e)
}

func (e ErrorType) Validate() error {
	switch e {
	case ErrorTypeInternal,
		ErrorTypeOutOfRange,
		ErrorTypeUnsupported,
		ErrorTypeNotFound,
		ErrorTypeAlreadyExists,
		ErrorTypeInvalidFormat,
		ErrorTypeFailedPrecondition,
		ErrorTypeNotImplemented:
		return nil
	default:
		return NewInvalidFormatError(fmt.Sprintf("invalid error type: %s", e))
	}
}
