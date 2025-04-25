package errorx

import (
	"fmt"
)

// AlreadyExistsErrorf creates a CliniaError with type ErrorTypeAlreadyExists and a formatted message
func AlreadyExistsErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeAlreadyExists,
		fmt.Sprintf(format, args...),
	)
}

func IsAlreadyExistsError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeAlreadyExists
}
