package logrusx

import (
	"testing"

	"github.com/clinia/x/errorx"
	"github.com/stretchr/testify/assert"
)

func TestCliniaErrorCtx(t *testing.T) {
	t.Run("should return error when no details", func(t *testing.T) {
		err := errorx.InvalidArgumentErrorf("invalid content")
		assert.Equal(t, map[string]interface{}{"message": "[INVALID_ARGUMENT] invalid content"}, cliniaErrorCtx(err))
	})

	t.Run("should return error with details", func(t *testing.T) {
		err := errorx.InvalidArgumentErrorf("invalid content")
		err = err.WithDetails(errorx.AlreadyExistsErrorf("field 'foo' already exists"))
		assert.Equal(t, map[string]any{
			"message": "[INVALID_ARGUMENT] invalid content",
			"details": []map[string]any{
				{
					"message": "[ALREADY_EXISTS] field 'foo' already exists",
				},
			},
		}, cliniaErrorCtx(err))
	})

	t.Run("should return error with nested details", func(t *testing.T) {
		err := errorx.InvalidArgumentErrorf("invalid content")
		err = err.WithDetails(errorx.AlreadyExistsErrorf("field 'foo' already exists"))
		nestedErr := errorx.InvalidArgumentErrorf("invalid field 'bar'")
		nestedErr = nestedErr.WithDetails(errorx.InvalidArgumentErrorf("missing field 'xyz'"))
		nestedErr = nestedErr.WithDetails(errorx.InvalidArgumentErrorf("additional field 'abc'"))
		err = err.WithDetails(nestedErr)

		assert.Equal(t, map[string]interface{}{
			"message": "[INVALID_ARGUMENT] invalid content",
			"details": []map[string]interface{}{
				{
					"message": "[ALREADY_EXISTS] field 'foo' already exists",
				},
				{
					"message": "[INVALID_ARGUMENT] invalid field 'bar'",
					"details": []map[string]interface{}{
						{
							"message": "[INVALID_ARGUMENT] missing field 'xyz'",
						},
						{
							"message": "[INVALID_ARGUMENT] additional field 'abc'",
						},
					},
				},
			},
		}, cliniaErrorCtx(err))
	})
}
