package errorx

import "fmt"

// UnimplementedErrorf creates a CliniaError with type ErrorTypeUnimplemented and a formatted message
func UnimplementedErrorf(format string, args ...interface{}) *CliniaError {
	return newWithStack(
		ErrorTypeUnimplemented,
		fmt.Sprintf(format, args...),
	)
}

func IsUnimplemented(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeUnimplemented
}
