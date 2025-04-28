package errorx

import "fmt"

// ContentTooLargeErrorf creates a CliniaError with type ErrorTypeContentTooLarge and a formatted message
func ContentTooLargeErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeContentTooLarge,
		fmt.Sprintf(format, args...),
	)
}

func IsContentTooLargeError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeContentTooLarge
}
