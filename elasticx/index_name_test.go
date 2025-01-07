package elasticx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexName(t *testing.T) {
	t.Run("should return a new index name", func(t *testing.T) {
		name := NewIndexName("a", "b", "c")
		assert.Equal(t, "a~b~c", name.String())
	})

	t.Run("should return the engine name", func(t *testing.T) {
		name := NewIndexName("a")
		assert.Equal(t, "", name.EngineName())

		name = NewIndexName("a", "b")
		assert.Equal(t, "b", name.EngineName())

		name = NewIndexName("a", "b", "c")
		assert.Equal(t, "b", name.EngineName())
	})

	t.Run("should return the index name", func(t *testing.T) {
		name := NewIndexName("a")
		assert.Equal(t, "", name.Name())

		name = NewIndexName("a", "b")
		assert.Equal(t, "", name.Name())

		name = NewIndexName("a", "b", "c")
		assert.Equal(t, "c", name.Name())

		name = NewIndexName("a", "b", "c", "d")
		assert.Equal(t, "c~d", name.Name())
	})
}
