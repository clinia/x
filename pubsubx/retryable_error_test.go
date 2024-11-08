package pubsubx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRetryableError(t *testing.T) {
	t.Run("return true and the retryable error object if the error is a retryable error", func(t *testing.T) {
		retryableError := NewRetryableError(errors.New("test"))
		err, ok := IsRetryableError(retryableError)
		assert.Equal(t, &retryableError, err)
		assert.True(t, ok)
	})

	t.Run("return false and the nil when the error is not a RetryableError", func(t *testing.T) {
		err, ok := IsRetryableError(errors.New("test"))
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return the proper message", func(t *testing.T) {
		err := NewRetryableError(errors.New("test"))
		assert.Equal(t, "Can retry: true - test", err.Error())
	})
}
