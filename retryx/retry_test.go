package retryx

import (
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		name          string
		fn            func() error
		opts          []RetryOption
		expectedCalls int
		expectedError error
	}{
		{
			name: "successful retry",
			fn: func() error {
				return nil
			},
			expectedCalls: 1,
			opts:          nil,
			expectedError: nil,
		},
		{
			name: "retry with permanent error",
			fn: func() error {
				return backoff.Permanent(errors.New("permanent error"))
			},
			expectedCalls: 1,
			opts:          nil,
			expectedError: errors.New("permanent error"),
		},
		{
			name: "retry with temporary error",
			fn: func() error {
				return errors.New("temporary error")
			},
			opts: []RetryOption{
				WithRetryCount(2),
			},
			expectedCalls: 2,
			expectedError: errors.New("temporary error"),
		},
		{
			name: "retry with custom initial interval",
			fn: func() error {
				return errors.New("temporary error")
			},
			opts: []RetryOption{
				WithInterval(100 * time.Millisecond),
				WithRetryCount(2),
			},
			expectedCalls: 2,
			expectedError: errors.New("temporary error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCalls := 0
			fn := func() error {
				actualCalls++
				return tt.fn()
			}
			err := ExponentialRetry(fn, tt.opts...)
			if tt.expectedError != nil {
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedCalls, actualCalls)
		})
	}
}

func TestQuickRetry(t *testing.T) {
	tests := []struct {
		name          string
		fn            func() error
		opts          []RetryOption
		expectedCalls int
		expectedError error
	}{
		{
			name: "successful retry",
			fn: func() error {
				return nil
			},
			expectedCalls: 1,
			opts:          nil,
			expectedError: nil,
		},
		{
			name: "retry with permanent error",
			fn: func() error {
				return backoff.Permanent(errors.New("permanent error"))
			},
			expectedCalls: 1,
			opts:          nil,
			expectedError: errors.New("permanent error"),
		},
		{
			name: "retry with temporary error",
			fn: func() error {
				return errors.New("temporary error")
			},
			opts: []RetryOption{
				WithRetryCount(2),
			},
			expectedCalls: 2,
			expectedError: errors.New("temporary error"),
		},
		{
			name: "retry with custom initial interval",
			fn: func() error {
				return errors.New("temporary error")
			},
			opts: []RetryOption{
				WithInterval(100 * time.Millisecond),
				WithRetryCount(2),
			},
			expectedCalls: 2,
			expectedError: errors.New("temporary error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCalls := 0
			fn := func() error {
				actualCalls++
				return tt.fn()
			}
			err := ExponentialRetry(fn, tt.opts...)
			if tt.expectedError != nil {
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedCalls, actualCalls)
		})
	}
}
