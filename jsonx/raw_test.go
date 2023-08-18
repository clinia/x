package jsonx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawMessage(t *testing.T) {
	t.Run("should normalize an indentented json string", func(t *testing.T) {
		raw := RawMessage(`{
			"foo": "bar"
		}`)

		assert.Equal(t, json.RawMessage(`{"foo":"bar"}`), raw)
	})
}
