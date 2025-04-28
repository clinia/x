package errorx

import "fmt"

// InternalErrorf creates a CliniaError with type ErrorTypeInternal and a formatted message
func InternalErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeInternal,
		fmt.Sprintf(format, args...),
	)
}

func IsInternalError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInternal
}
