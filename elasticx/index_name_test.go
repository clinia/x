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
}
