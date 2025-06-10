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
