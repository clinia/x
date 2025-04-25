package errorx

import "fmt"

// UnauthenticatedErrorf creates a CliniaError with type ErrorTypeUnauthenticated and a formatted message
func UnauthenticatedErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeUnauthenticated,
		fmt.Sprintf(format, args...),
	)
}

func IsUnauthenticatedError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeUnauthenticated
}
