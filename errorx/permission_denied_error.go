package errorx

import "fmt"

// PermissionDeniedErrorf creates a CliniaError with type ErrorTypePermissionDenied and a formatted message
func PermissionDeniedErrorf(format string, args ...interface{}) *CliniaError {
	return newWithStack(
		ErrorTypePermissionDenied,
		fmt.Sprintf(format, args...),
	)
}

func IsPermissionDeniedError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypePermissionDenied
}
