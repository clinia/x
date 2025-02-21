package retryx

import "time"

type retryOptions struct {
	retryCount      int
	initialInterval time.Duration
}

type RetryOption func(*retryOptions)

func WithRetryCount(count int) RetryOption {
	return func(ro *retryOptions) {
		ro.retryCount = count
	}
}

func WithInitialDuration(interval time.Duration) RetryOption {
	return func(ro *retryOptions) {
		ro.initialInterval = interval
	}
}
