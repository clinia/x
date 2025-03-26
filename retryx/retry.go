package retryx

import (
	"time"

	"github.com/cenkalti/backoff"
)

const (
	DefaultInterval       = 500 * time.Millisecond
	DefaultMaxInterval    = 2 * time.Second
	DefaultMaxElapsedTime = 5 * time.Second
	DefaultMaxRetries     = 3
)

// ConstantRetry executes the provided function `fn` with a constant retry interval.
// This function is designed for simple retry scenarios where the interval between retries
// and the maximum number of retries are the only customizable options.
//
// Parameters:
// - fn: The function to be executed and retried upon failure.
// - opts: Optional retry configurations, such as initial interval and maximum retries.
//
// The retry interval defaults to `DefaultInterval` unless overridden by the `WithInterval`
// option. If more advanced control over the retry behavior is required, consider using the
// `backoff` package directly.
func ConstantRetry(fn func() error, opts ...RetryOption) error {
	rOpts := &retryOptions{}
	for _, opt := range opts {
		opt(rOpts)
	}

	duration := DefaultInterval
	if rOpts.initialInterval > 0 {
		duration = rOpts.initialInterval
	}

	bc := backoff.NewConstantBackOff(duration)
	bc.Reset()

	return retry(fn, bc, rOpts)
}

// ExponentialRetry executes the provided function `fn` with an exponential backoff retry strategy.
// This function is suitable for scenarios where the retry interval increases exponentially
// with each attempt, up to a maximum interval and elapsed time.
//
// Parameters:
// - fn: The function to be executed and retried upon failure.
// - opts: Optional retry configurations, such as initial interval, maximum retries, and other settings.
//
// The retry interval starts at `DefaultInterval` unless overridden by the `WithInterval` option.
// The maximum interval between retries starts at `DefaultMaxInterval` unless overridden by the `WithMaxInterval` option.
// The maximum elapsed time defaults to `DefaultMaxElapsedTime` unless overridden by the `WithMaxElapsedTime`option.
// If more advanced control over the retry behavior is required, consider using the `backoff` package directly.
func ExponentialRetry(fn func() error, opts ...RetryOption) error {
	rOpts := &retryOptions{}
	for _, opt := range opts {
		opt(rOpts)
	}

	duration := DefaultInterval
	maxInterval := DefaultMaxInterval
	maxElapsedTime := DefaultMaxElapsedTime
	if rOpts.initialInterval > 0 {
		duration = rOpts.initialInterval
	}
	if rOpts.maxInterval > 0 {
		maxInterval = rOpts.maxInterval
	}
	if rOpts.maxElapsedTime > 0 {
		maxElapsedTime = rOpts.maxElapsedTime
	}

	bc := backoff.NewExponentialBackOff()
	bc.InitialInterval = duration
	bc.MaxInterval = maxInterval
	bc.MaxElapsedTime = maxElapsedTime
	bc.Reset()

	return retry(fn, bc, rOpts)
}

func retry(fn func() error, bo backoff.BackOff, rOpts *retryOptions) error {
	maxRetryCount := DefaultMaxRetries
	if rOpts.retryCount > 0 {
		maxRetryCount = rOpts.retryCount
	}

	retries := 0
	return backoff.Retry(func() error {
		err := fn()
		if err == nil {
			return nil
		}

		retries++
		if retries >= maxRetryCount {
			return backoff.Permanent(err)
		}

		return err
	}, bo)
}
