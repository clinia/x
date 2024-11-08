package pubsubx

import "fmt"

type RetryableError struct {
	OriginalError error
	Retryable     bool
}

func NewRetryableError(err error) RetryableError {
	return RetryableError{
		OriginalError: err,
		Retryable:     true,
	}
}

func (re RetryableError) Error() string {
	return fmt.Sprintf("Can retry: %v - %s", re.Retryable, re.OriginalError.Error())
}

func IsRetryableError(err error) (*RetryableError, bool) {
	re, ok := err.(RetryableError)
	if !ok {
		return nil, false
	}
	return &re, true
}
