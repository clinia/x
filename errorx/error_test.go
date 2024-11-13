package errorx

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("should return clinia error from stack", func(t *testing.T) {
		err := NewAlreadyExistsError("test")
		serr := errors.WithStack(err)

		_, ok := IsCliniaError(serr)
		assert.True(t, ok)
	})

	t.Run("should return a clinia error without stack", func(t *testing.T) {
		err := NewAlreadyExistsError("test")

		_, ok := IsCliniaError(err)
		assert.True(t, ok)
	})

	t.Run("should return is not found from stack", func(t *testing.T) {
		err := errors.WithStack(NotFoundErrorf("test"))
		assert.True(t, IsNotFoundError(err))
	})

	t.Run("should return is not found", func(t *testing.T) {
		err := NotFoundErrorf("test")
		assert.True(t, IsNotFoundError(err))
	})

	t.Run("should return as RetryableError", func(t *testing.T) {
		err := NotFoundErrorf("test")
		retryErr := err.AsRetryableError()
		assert.Equal(t, retryErr.Unwrap(), err)
	})
}
