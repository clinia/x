package retryx_test

import (
	"time"

	"github.com/clinia/x/retryx"
)

func ExampleConstantRetry() {
	// Define a function to be retried.
	fn := func() error {
		// Perform some operation that may fail.
		return nil
	}

	// Retry the function with default options.
	err := retryx.ConstantRetry(fn)
	if err != nil {
		// Handle the error.
	}

	// Retry the function with custom options.
	err = retryx.ConstantRetry(fn, retryx.WithRetryCount(5), retryx.WithInterval(1*time.Second))
	if err != nil {
		// Handle the error.
	}
}

func ExampleExponentialRetry() {
	// Define a function to be retried.
	fn := func() error {
		// Perform some operation that may fail.
		return nil
	}

	// Retry the function with default options.
	err := retryx.ExponentialRetry(fn)
	if err != nil {
		// Handle the error.
	}

	// Retry the function with custom options.
	err = retryx.ExponentialRetry(fn,
		retryx.WithMaxElapsedTime(10*time.Second),
		retryx.WithMaxInterval(2*time.Second),
		retryx.WithInterval(50*time.Millisecond),
		retryx.WithRetryCount(5),
	)
	if err != nil {
		// Handle the error.
	}
}
