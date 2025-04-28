package errorx

import "fmt"

// InvalidArgumentErrorf creates a CliniaError with type ErrorTypeInvalidArgument and a formatted message
func InvalidArgumentErrorf(format string, args ...any) *CliniaError {
	return newWithStack(
		ErrorTypeInvalidArgument,
		fmt.Sprintf(format, args...),
	)
}

func IsInvalidArgumentError(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeInvalidArgument
}
