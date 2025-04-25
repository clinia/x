package errorx

import "fmt"

// FailedPreconditionErrorf creates a CliniaError with type ErrorTypeFailedPrecondition and a formatted message
func FailedPreconditionErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeFailedPrecondition,
		fmt.Sprintf(format, args...),
	)
}

func IsFailedPreconditionError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeFailedPrecondition
}
