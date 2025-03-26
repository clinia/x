package retryx

import "time"

type (
	retryOptions struct {
		retryCount      int
		initialInterval time.Duration
		maxInterval     time.Duration
		maxElapsedTime  time.Duration
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

// WithMaxInterval sets the maximum interval between retries.
// Default max interval is 2s.
// When used with ExponentialRetry, this sets the maximum interval.
// When used with ConstantRetry, this option is ignored.
func WithMaxInterval(maxInterval time.Duration) RetryOption {
	return func(ro *retryOptions) {
		ro.maxInterval = maxInterval
	}
}

// WithMaxElapsedTime sets the maximum elapsed time for retries.
// Default max elapsed time is 5s.
// When used with ExponentialRetry, this sets the maximum elapsed time.
// When used with ConstantRetry, this option is ignored.
func WithMaxElapsedTime(maxElapsedTime time.Duration) RetryOption {
	return func(ro *retryOptions) {
		ro.maxElapsedTime = maxElapsedTime
	}
}
