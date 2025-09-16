package errorx

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/clinia/x/pointerx"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type CliniaError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`

	// Not returned to clients
	OriginalError error
	// List of errors that caused the error if applicable
	Details []CliniaError
}

var _ error = (*CliniaError)(nil)

type CliniaErrors []*CliniaError

func (cErrs CliniaErrors) AsErrors() []error {
	errs := make([]error, len(cErrs))
	for i, cErr := range cErrs {
		if cErr != nil {
			errs[i] = error(cErr)
		}
	}
	return errs
}

func CliniaErrorsFromCliniaErrorSlice(errs []CliniaError) CliniaErrors {
	cErrs := make(CliniaErrors, len(errs))
	for i, err := range errs {
		if err.Type == "" && err.Message == "" && err.OriginalError == nil && len(err.Details) == 0 {
			continue
		}
		cErrs[i] = &err
	}
	return cErrs
}

func CliniaErrorsFromErrorSlice(errs []error) CliniaErrors {
	cErrs := make(CliniaErrors, len(errs))
	for i, err := range errs {
		if err == nil {
			continue
		}
		var ok bool
		cErrs[i], ok = IsCliniaError(err)
		if !ok {
			var mErr error
			cErrs[i], mErr = NewCliniaErrorFromMessage(err.Error())
			if mErr != nil {
				cErrs[i] = pointerx.Ptr(NewCliniaInternalErrorFromError(err))
			}
		}
	}
	return cErrs
}

func (e CliniaError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// WithDetails allows to attach multiple errors to the error
func (e *CliniaError) WithDetails(details ...CliniaError) CliniaError {
	if len(details) == 0 {
		return *e
	}
	if e.Details == nil {
		e.Details = make([]CliniaError, 0, len(details))
	}
	e.Details = append(e.Details, details...)
	return *e
}

// WithDetails allows to attach multiple errors to the error
func (e *CliniaError) WithErrorDetails(details ...error) CliniaError {
	if len(details) == 0 {
		return *e
	}
	cDetails := lo.Map(details, func(item error, index int) CliniaError {
		cErr, ok := IsCliniaError(item)
		if !ok {
			cErr, _ = NewCliniaErrorFromMessage(item.Error())
			if cErr == nil {
				cErr = pointerx.Ptr(NewCliniaInternalErrorFromError(item))
			}
		}
		return *cErr
	})
	return e.WithDetails(cDetails...)
}

func (e *CliniaError) AsRetryableError() RetryableError {
	return NewRetryableError(*e)
}

func (e *CliniaError) IsRetryable() bool {
	_, ok := IsRetryableError(e.OriginalError)
	return ok
}

func NewCliniaInternalErrorFromError(err error) CliniaError {
	cErr := InternalErrorf(err.Error())
	cErr.OriginalError = err
	return cErr
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
	mE, ok := e.(*CliniaError)
	if !ok {
		mEs, ok := e.(CliniaError)
		if !ok {
			return nil, false
		}
		mE = &mEs
	}

	if mE.Type == ErrorTypeUnspecified {
		return nil, false
	}

	return mE, true
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

func IsContentTooLargeError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeContentTooLarge
}

// AlreadyExistsErrorf creates a CliniaError with type ErrorTypeAlreadyExists and a formatted message
func AlreadyExistsErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: fmt.Sprintf(format, args...),
	}
}

// AlreadyExistsError creates a CliniaError with type ErrorTypeAlreadyExists and a message
func AlreadyExistsError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeAlreadyExists,
		Message: error,
	}
}

// FailedPreconditionErrorf creates a CliniaError with type ErrorTypeFailedPrecondition and a formatted message
func FailedPreconditionErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeFailedPrecondition,
		Message: fmt.Sprintf(format, args...),
	}
}

// FailedPreconditionError creates a CliniaError with type ErrorTypeFailedPrecondition and a message
func FailedPreconditionError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeFailedPrecondition,
		Message: error,
	}
}

// InternalErrorf creates a CliniaError with type ErrorTypeInternal and a formatted message
func InternalErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInternal,
		Message: fmt.Sprintf(format, args...),
	}
}

func InternalError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInternal,
		Message: error,
	}
}

// InvalidArgumentErrorf creates a CliniaError with type ErrorTypeInvalidArgument and a formatted message
func InvalidArgumentErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidArgument,
		Message: fmt.Sprintf(format, args...),
	}
}

// InvalidArgumentError creates a CliniaError with type ErrorTypeInvalidArgument and a message
func InvalidArgumentError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeInvalidArgument,
		Message: error,
	}
}

// NotFoundErrorf creates a CliniaError with type ErrorTypeNotFound and a formatted message
func NotFoundErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// NotFoundError creates a CliniaError with type ErrorTypeNotFound and a message
func NotFoundError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeNotFound,
		Message: error,
	}
}

// OutOfRangeErrorf creates a CliniaError with type ErrorTypeOutOfRange and a formatted message
func OutOfRangeErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: fmt.Sprintf(format, args...),
	}
}

// OutOfRangeError creates a CliniaError with type ErrorTypeOutOfRange and a message
func OutOfRangeError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeOutOfRange,
		Message: error,
	}
}

// UnimplementedErrorf creates a CliniaError with type ErrorTypeUnimplemented and a formatted message
func UnimplementedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnimplemented,
		Message: fmt.Sprintf(format, args...),
	}
}

// UnimplementedError creates a CliniaError with type ErrorTypeUnimplemented and a message
func UnimplementedError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnimplemented,
		Message: error,
	}
}

// UnauthenticatedErrorf creates a CliniaError with type ErrorTypeUnauthenticated and a formatted message
func UnauthenticatedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnauthenticated,
		Message: fmt.Sprintf(format, args...),
	}
}

// UnauthenticatedError creates a CliniaError with type ErrorTypeUnauthenticated and a message
func UnauthenticatedError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeUnauthenticated,
		Message: error,
	}
}

// PermissionDeniedErrorf creates a CliniaError with type ErrorTypePermissionDenied and a formatted message
func PermissionDeniedErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypePermissionDenied,
		Message: fmt.Sprintf(format, args...),
	}
}

// PermissionDeniedError creates a CliniaError with type ErrorTypePermissionDenied and a message
func PermissionDeniedError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypePermissionDenied,
		Message: error,
	}
}

// ContentTooLargeErrorf creates a CliniaError with type ErrorTypeContentTooLarge and a formatted message
func ContentTooLargeErrorf(format string, args ...interface{}) CliniaError {
	return CliniaError{
		Type:    ErrorTypeContentTooLarge,
		Message: fmt.Sprintf(format, args...),
	}
}

// ContentTooLargeError creates a CliniaError with type ErrorTypeContentTooLarge and a message
func ContentTooLargeError(error string) CliniaError {
	return CliniaError{
		Type:    ErrorTypeContentTooLarge,
		Message: error,
	}
}

// Deprecated: use InternalErrorf instead
func NewInternalError(e error) CliniaError {
	return CliniaError{
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
