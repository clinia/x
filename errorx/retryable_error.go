package errorx

import "fmt"

type RetryableError struct {
	error
}

var _ error = (*RetryableError)(nil)

func NewRetryableError(err error) RetryableError {
	return RetryableError{
		error: err,
	}
}

func (re RetryableError) Unwrap() error {
	return re.error
}

func (re RetryableError) Error() string {
	return fmt.Sprintf("Retryable - %s", re.error.Error())
}

func IsRetryableError(err error) (*RetryableError, bool) {
	re, ok := err.(RetryableError)
	if !ok {
		return nil, false
	}
	return &re, true
}
