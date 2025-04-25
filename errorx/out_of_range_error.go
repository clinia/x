package errorx

import "fmt"

// OutOfRangeErrorf creates a CliniaError with type ErrorTypeOutOfRange and a formatted message
func OutOfRangeErrorf(format string, args ...interface{}) *CliniaError {
	return newWithStack(
		ErrorTypeOutOfRange,
		fmt.Sprintf(format, args...),
	)
}

func IsOutOfRange(e error) bool {
	mE, ok := IsCliniaError(e)
	if !ok {
		return false
	}

	return mE.Type == ErrorTypeOutOfRange
}
