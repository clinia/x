package errorx

type ErrorType string

// Errors status code are defined here:
// https://chromium.googlesource.com/external/github.com/grpc/grpc/+/refs/tags/v1.21.4-pre1/doc/statuscodes.md

const (
	// The Invalid type should not be used, only useful to assert whether or not an error is an MdmError during cast
	ErrorTypeUnspecified        = ErrorType("")
	ErrorTypeAlreadyExists      = ErrorType("ALREADY_EXISTS")
	ErrorTypeFailedPrecondition = ErrorType("FAILED_PRECONDITION")
	ErrorTypeInternal           = ErrorType("INTERNAL")
	ErrorTypeInvalidArgument    = ErrorType("INVALID_ARGUMENT")
	ErrorTypeNotFound           = ErrorType("NOT_FOUND")
	ErrorTypeOutOfRange         = ErrorType("OUT_OF_RANGE")
	ErrorTypeUnimplemented      = ErrorType("UNIMPLEMENTED")
	ErrorTypeUnauthenticated    = ErrorType("UNAUTHENTICATED")
	ErrorTypePermissionDenied   = ErrorType("PERMISSION_DENIED")
	ErrorTypePayloadTooLarge    = ErrorType("PAYLOAD_TOO_LARGE")
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
	case ErrorTypeAlreadyExists,
		ErrorTypeFailedPrecondition,
		ErrorTypeInternal,
		ErrorTypeInvalidArgument,
		ErrorTypeNotFound,
		ErrorTypeOutOfRange,
		ErrorTypeUnimplemented,
		ErrorTypeUnauthenticated,
		ErrorTypePermissionDenied,
		ErrorTypePayloadTooLarge:
		return nil
	default:
		return InvalidArgumentErrorf("invalid error type: %s", e)
	}
}
