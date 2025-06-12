package utilx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	errFunc := func() (*struct{}, error) {
		return nil, errors.New("my error")
	}
	returnFunc := func() (*struct{}, error) {
		return &struct{}{}, nil
	}
	t.Run("should panic when error returned", func(t *testing.T) {
		assert.Panics(t, func() {
			Must(errFunc())
		})
	})
	t.Run("should return value when no error", func(t *testing.T) {
		assert.NotNil(t, Must(returnFunc()))
	})
}
