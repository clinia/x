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
		assert.Equal(t, "a", NewIndexName("a").EngineName())
		assert.Equal(t, "b", NewIndexName(enginesIndexNameSegment, "b").EngineName())
		assert.Equal(t, "b", NewIndexName(enginesIndexNameSegment, "b", "c").EngineName())
		assert.Equal(t, "", IndexName("").EngineName())
	})

	t.Run("should return the index name", func(t *testing.T) {
		name := NewIndexName("a")
		assert.Equal(t, "", name.Name())

		name = NewIndexName("a", "b")
		assert.Equal(t, "", name.Name())

		name = NewIndexName("a", "b", "c")
		assert.Equal(t, "b~c", name.Name())

		name = NewIndexName(enginesIndexNameSegment, "b", "c", "d")
		assert.Equal(t, "c~d", name.Name())
	})
}

func TestFullIndexName(t *testing.T) {
	t.Run("should return a new index name", func(t *testing.T) {
		name := FullIndexName("a", "b", "c")
		assert.Equal(t, "clinia-engines~a~b~c", name.String())
	})

	t.Run("should return the engine name", func(t *testing.T) {
		assert.Equal(t, "a", FullIndexName("a").EngineName())
		assert.Equal(t, "a", FullIndexName("a", "b").EngineName())
		assert.Equal(t, "a", FullIndexName("a", "b", "c").EngineName())
	})

	t.Run("should return the index name", func(t *testing.T) {
		name := FullIndexName("a")
		assert.Equal(t, "", name.Name())

		name = FullIndexName("a", "b")
		assert.Equal(t, "b", name.Name())

		name = FullIndexName("a", "b", "c")
		assert.Equal(t, "b~c", name.Name())
	})

}
