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

	t.Run("should return a RetryableError wrapped in a CliniaError without OriginalError", func(t *testing.T) {
		cerr := CliniaError{Type: ErrorTypeNotFound, Message: "Test"}
		assert.False(t, cerr.IsRetryable())
		cerr = cerr.AsRetryableError()
		assert.True(t, cerr.IsRetryable())
	})

	t.Run("should return a RetryableError wrapped in a CliniaError with OriginalError", func(t *testing.T) {
		cerr := NewInternalError(errors.New("test"))
		assert.False(t, cerr.IsRetryable())
		cerr = cerr.AsRetryableError()
		assert.True(t, cerr.IsRetryable())
	})

	t.Run("should append details to existing error", func(t *testing.T) {
		cerr := FailedPreconditionErrorf("test")
		cerr = cerr.WithDetails(NotFoundErrorf("testnotfound"))
		assert.Equal(t, CliniaError{
			Type:    ErrorTypeFailedPrecondition,
			Message: "test",
			Details: []CliniaError{
				{
					Type:    ErrorTypeNotFound,
					Message: "testnotfound",
				},
			},
		}, cerr)

		// Append more details
		cerr = cerr.WithDetails(InvalidArgumentErrorf("testinvalid"))
		assert.Equal(t, CliniaError{
			Type:    ErrorTypeFailedPrecondition,
			Message: "test",
			Details: []CliniaError{
				{
					Type:    ErrorTypeNotFound,
					Message: "testnotfound",
				},
				{
					Type:    ErrorTypeInvalidArgument,
					Message: "testinvalid",
				},
			},
		}, cerr)
	})
}
