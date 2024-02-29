package errorx

import (
	"fmt"
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

func IsAlreadyExistsError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeAlreadyExists
}

func IsFailedPreconditionError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeFailedPrecondition
}

func IsInternalError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInternal
}

func IsInvalidArgumentError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInvalidArgument
}

func IsNotFoundError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeNotFound
}

func IsOutOfRange(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeOutOfRange
}

func IsUnimplemented(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeUnimplemented
}

func IsUnauthenticatedError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeUnauthenticated
}

func IsPermissionDeniedError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypePermissionDenied
}

// AlreadyExistsErrorf creates a CliniaError with type ErrorTypeAlreadyExists and a formatted message
func AlreadyExistsErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
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

// InternalErrorf creates a CliniaError with type ErrorTypeInternal and a formatted message
func InternalErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf(format, args...),
	}
}

// InvalidArgumentErrorf creates a CliniaError with type ErrorTypeInvalidArgument and a formatted message
func InvalidArgumentErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidArgument,
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

// OutOfRangeErrorf creates a CliniaError with type ErrorTypeOutOfRange and a formatted message
func OutOfRangeErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf(format, args...),
	}
}

// UnimplementedErrorf creates a CliniaError with type ErrorTypeUnimplemented and a formatted message
func UnimplementedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnimplemented,
		Message: fmt.Sprintf(format, args...),
	}
}

// UnauthenticatedErrorf creates a CliniaError with type ErrorTypeUnauthenticated and a formatted message
func UnauthenticatedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnauthenticated,
		Message: fmt.Sprintf(format, args...),
	}
}

// PermissionDeniedErrorf creates a CliniaError with type ErrorTypePermissionDenied and a formatted message
func PermissionDeniedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypePermissionDenied,
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

// Deprecated: use AlreadyExistsErrorf instead
func NewAlreadyExistsError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: message,
	}
}
