package retryx

import (
	"time"

	"github.com/cenkalti/backoff"
)

const (
	DefaultInterval   = 500 * time.Millisecond
	DefaultMaxRetries = 3
)

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

func ExponentialRetry(fn func() error, opts ...RetryOption) error {
	rOpts := &retryOptions{}
	for _, opt := range opts {
		opt(rOpts)
	}

	duration := DefaultInterval
	if rOpts.initialInterval > 0 {
		duration = rOpts.initialInterval
	}

	bc := backoff.NewExponentialBackOff()
	bc.InitialInterval = duration
	bc.MaxInterval = 2 * time.Second
	bc.MaxElapsedTime = time.Second * 5
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
