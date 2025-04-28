package errorx

import "fmt"

// NotFoundErrorf creates a CliniaError with type ErrorTypeNotFound and a formatted message
func NotFoundErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeNotFound,
		fmt.Sprintf(format, args...),
	)
}

func IsNotFoundError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeNotFound
}
