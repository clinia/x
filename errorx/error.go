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

func NewUnsupportedError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnsupported,
		Message: message,
	}
}

func NewInvalidFormatError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidFormat,
		Message: message,
	}
}

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

func NewAlreadyExistsError(message string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: message,
	}
}
