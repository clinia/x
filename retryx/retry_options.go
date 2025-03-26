package retryx

import "time"

type (
	retryOptions struct {
		retryCount      int
		initialInterval time.Duration
	}

	RetryOption func(*retryOptions)
)

// WithRetryCount sets the maximum number of retries.
// Default retry count is 3.
// A retry count of <= 0 will not be taken into account (default will be used).
//
// Note: a retry count of 1 will result in 2 calls (the first one and 1 retry).
func WithRetryCount(count int) RetryOption {
	return func(ro *retryOptions) {
		ro.retryCount = count
	}
}

// WithInterval sets the initial interval between retries.
// Default interval is 500ms.
// When used with ExponentialRetry, this sets the initial interval.
// When used with ConstantRetry, this sets the constant interval.
func WithInterval(interval time.Duration) RetryOption {
	return func(ro *retryOptions) {
		ro.initialInterval = interval
	}
}
