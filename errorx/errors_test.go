package errorx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputErrsMatchInputLength(t *testing.T) {
	t.Run("should return an error on length mismatch", func(t *testing.T) {
		err := OutputErrsMatchInputLength(0, 2, nil)
		require.Error(t, err)
		assert.Equal(t, "[INTERNAL] a different length of errors (0) then the input length (2) was returned", err.Error())
	})

	t.Run("should return an error on length mismatch with error", func(t *testing.T) {
		err := OutputErrsMatchInputLength(0, 2, InvalidArgumentErrorf("test error"))
		require.Error(t, err)
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
	})
	t.Run("should return no error on length match", func(t *testing.T) {
		assert.NoError(t, OutputErrsMatchInputLength(2, 2, nil))
	})

	t.Run("should return error on length match with error", func(t *testing.T) {
		err := OutputErrsMatchInputLength(2, 2, InvalidArgumentErrorf("test error"))
		require.Error(t, err)
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
	})
}
