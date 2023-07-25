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
}
