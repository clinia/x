package errorx

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Wrapper provides context around another error.
type Wrapper interface {
	error
	Unwrap() error
}

// StackTracer provides a stack trace for an error.
type StackTracer interface {
	error
	StackTrace() Callers
}

// DetailedError provides extended information about an error.
// The ErrorDetails method returns a longer, multi-line description of
// the error. It always ends with a new line.
type DetailedError interface {
	error
	ErrorDetails() string
}

type CliniaError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`

	// List of errors that caused the error if applicable
	Details []*CliniaError

	err   error
	stack Callers
}

var _ error = (*CliniaError)(nil)
var _ StackTracer = (*CliniaError)(nil)
var _ DetailedError = (*CliniaError)(nil)

// ErrorDetails implements DetailedError.
func (e *CliniaError) ErrorDetails() string {
	return e.stack.String()
}

// StackTrace implements StackTracer.
func (e *CliniaError) StackTrace() Callers {
	return e.stack
}

func (e CliniaError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// WithDetails allows to attach multiple errors to the error
func (c *CliniaError) WithDetails(details ...error) *CliniaError {
	if c.Details == nil {
		c.Details = make([]*CliniaError, 0, len(details))
	}

	for _, detail := range details {
		if d, ok := IsCliniaError(detail); ok {
			c.Details = append(c.Details, d)
		}
	}

	return c
}

func (c *CliniaError) AsRetryableError() RetryableError {
	return NewRetryableError(*c)
}

func (c *CliniaError) IsRetryable() bool {
	_, ok := IsRetryableError(c.err)
	return ok
}

// newWithStack Create a new CliniaError with the given error type and message
// The stack trace is captured at the point of creation
func newWithStack(eType ErrorType, message string) *CliniaError {
	return &CliniaError{
		Type:    eType,
		Message: message,
		stack:   callers(2),
	}
}

func NewCliniaErrorFromMessage(msg string) (*CliniaError, error) {
	r, _ := regexp.Compile(`\[(.*?)\] (.*)`)
	m := r.FindStringSubmatch(msg)
	if len(m) < 2 {
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

	if mE, ok := e.(CliniaError); ok {
		if mE.Type == ErrorTypeUnspecified {
			return nil, false
		}
		return &mE, true
	}

	if mE, ok := e.(*CliniaError); ok {
		if mE.Type == ErrorTypeUnspecified {
			return nil, false
		}
		return mE, true
	}

	return nil, false
}

func NewEnumOutOfRangeError(actual string, expectedOneOf []string, enumName string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf("%q is not a valid %s. Possible values: [%s]", actual, enumName, strings.Join(expectedOneOf, ", ")),
	}
}
