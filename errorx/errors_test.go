package errorx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputErrsMatchInputLength(t *testing.T) {
	t.Run("should return an error on length mismatch", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{}, 2, nil)
		assert.Len(t, errs, 2)
		require.Error(t, err)
		assert.Equal(t, "[INTERNAL] a different length of errors (0) then the input length (2) was returned", err.Error())
	})

	t.Run("should return an error on length mismatch with error", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{}, 2, InvalidArgumentErrorf("test error"))
		assert.Len(t, errs, 2)
		require.Error(t, err)
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
	})
	t.Run("should return no error on length match", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{nil, errors.New("test")}, 2, nil)
		assert.NoError(t, err)
		assert.Len(t, errs, 2)
	})

	t.Run("should return error on length match with error", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{nil, errors.New("test")}, 2, InvalidArgumentErrorf("test error"))
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
		require.Error(t, err)
		assert.Len(t, errs, 2)
	})
}

func TestErrorWithAdditionalContext(t *testing.T) {
	t.Run("should be able to add additional context to an error", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithAdditionalContext("test", "ing").
			WithAdditionalContext("numerical", 10).
			WithApplicationErrorCode("TEST_CODE")

		assert.Equal(t, map[string]any{
			"test":      "ing",
			"numerical": 10,
		}, err.AdditionalContext)
		assert.Equal(t, "TEST_CODE", err.ApplicationErrorCode)
	})

	t.Run("should only take latest additional context value for a given key", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithAdditionalContext("test", "ing").
			WithAdditionalContext("test", "ed")

		assert.Equal(t, map[string]any{
			"test": "ed",
		}, err.AdditionalContext)
	})

	t.Run("should only take latest application error code", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithApplicationErrorCode("TEST_CODE").
			WithApplicationErrorCode("TEST_CODE_2")

		assert.Equal(t, "TEST_CODE_2", err.ApplicationErrorCode)
	})

	t.Run("should be able to attach additional details on an error pointer", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error")
		errPtr := &err
		err = errPtr.WithAdditionalContext("test", "ing").
			WithAdditionalContext("numerical", 10).
			WithApplicationErrorCode("TEST_CODE")

		assert.Equal(t, map[string]any{
			"test":      "ing",
			"numerical": 10,
		}, err.AdditionalContext)
		assert.Equal(t, "TEST_CODE", err.ApplicationErrorCode)
	})

	t.Run("should be able to attach additional context using maps", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithAdditionalContextMap(map[string]any{
				"test":      "ing",
				"numerical": 10,
				"overriden": "old",
			}, map[string]any{
				"overriden": "new",
			}).
			WithApplicationErrorCode("TEST_CODE")

		assert.Equal(t, map[string]any{
			"test":      "ing",
			"numerical": 10,
			"overriden": "new",
		}, err.AdditionalContext)
		assert.Equal(t, "TEST_CODE", err.ApplicationErrorCode)
	})
}
