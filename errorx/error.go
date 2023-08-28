package errorx

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type CliniaError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`

	OriginalError error // Not returned to clients
}

var _ error = (*CliniaError)(nil)

type CliniaRetryableError = CliniaError

func (e CliniaError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

func NewCliniaErrorFromMessage(msg string) (*CliniaError, error) {
	r, _ := regexp.Compile(`\[(.*?)\] (.*)`)
	m := r.FindStringSubmatch(msg)
	if m == nil || len(m) < 2 {
		return nil, fmt.Errorf("%q is not a valid error type", msg)
	}

	eT, err := ParseErrorType(m[1])
	if err != nil {
		return nil, err
	}

	if len(m) >= 3 {
		msg = m[2]
	}

	return &CliniaError{
		Type:    eT,
		Message: msg,
	}, nil
}

func IsCliniaError(e error) (*CliniaError, bool) {
	e = errors.Cause(e)
	mE, ok := e.(CliniaError)
	if !ok {
		return nil, false
	}

	if mE.Type == ErrorTypeUnspecified {
		return nil, false
	}

	return &mE, true
}

func IsNotFoundError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeNotFound
}

func IsFailedPreconditionError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeFailedPrecondition
}

func IsAlreadyExistsError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeAlreadyExists
}

func IsInvalidFormatError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInvalidFormat
}

func IsUnsupportedError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeUnsupported
}

func IsInternalError(e error) bool {
	mE, ok := e.(CliniaError)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInternal
}

// InternalErrorf creates a CliniaError with type ErrorTypeInternal and a formatted message
func InternalErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf(format, args...),
	}
}

// NotFoundErrorf creates a CliniaError with type ErrorTypeNotFound and a formatted message
func NotFoundErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// UnsupportedErrorf creates a CliniaError with type ErrorTypeUnsupported and a formatted message
func UnsupportedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnsupported,
		Message: fmt.Sprintf(format, args...),
	}
}

// InvalidFormatErrorf creates a CliniaError with type ErrorTypeInvalidFormat and a formatted message
func InvalidFormatErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidFormat,
		Message: fmt.Sprintf(format, args...),
	}
}

// FailedPreconditionErrorf creates a CliniaError with type ErrorTypeFailedPrecondition and a formatted message
func FailedPreconditionErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeFailedPrecondition,
		Message: fmt.Sprintf(format, args...),
	}
}

// AlreadyExistsErrorf creates a CliniaError with type ErrorTypeAlreadyExists and a formatted message
func AlreadyExistsErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: fmt.Sprintf(format, args...),
	}
}

// Deprecated: use InternalErrorf instead
func NewInternalError(e error) CliniaRetryableError {
	return CliniaRetryableError{
		Type:          ErrorTypeInternal,
		Message:       "Internal Error",
		OriginalError: e,
	}
}

func NewEnumOutOfRangeError(actual string, expectedOneOf []string, enumName string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf("%q is not a valid %s. Possible values: [%s]", actual, enumName, strings.Join(expectedOneOf, ", ")),
	}
}

// Deprecated: use NotFoundErrorf instead
func NewNotFoundError(msg string, err error) CliniaError {
	x := "Resource"
	if msg != "" {
		x = msg
	}

	message := fmt.Sprintf("%s could not be found", x)

	return CliniaError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

// Deprecated: use UnsupportedErrorf instead
func NewUnsupportedError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnsupported,
		Message: message,
	}
}

// Deprecated: use UnsupportedErrorf instead
func NewInvalidFormatError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidFormat,
		Message: message,
	}
}

// Deprecated: use FailedPreconditionErrorf instead
func NewFailedPreconditionError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeFailedPrecondition,
		Message: message,
	}
}

func NewSingleOutOfRangeError(actual string, expectedOneOf []string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf("%s is not in possible values: [%s]", actual, strings.Join(expectedOneOf, ", ")),
	}
}

func NewMultipleOutOfRangeError(actual []string, expectedOneOf []string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf("%s are not in possible values: [%s]", strings.Join(actual, ", "), strings.Join(expectedOneOf, ", ")),
	}
}

func NewUnsupportedValueError(actual string, expected string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnsupported,
		Message: fmt.Sprintf("Expected value %s but got %s", expected, actual),
	}
}

func NewUnsupportedTypeError[T1 any, T2 any](actual T1, expected T2) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnsupported,
		Message: fmt.Sprintf("Expected type %s but got %s", reflect.TypeOf(expected), reflect.TypeOf(actual)),
	}
}

// Deprecated: use AlreadyExistsErrorf instead
func NewAlreadyExistsError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: message,
	}
}

// Deprecated: use InternalErrorf instead
func NewNotImplementedError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeNotImplemented,
		Message: message,
	}
}
