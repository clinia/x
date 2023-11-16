package slicex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	original := []string{"foo", "bar"}
	new := Append(original, "baz")

	assert.Equal(t, []string{"foo", "bar", "baz"}, new)

	// Ensure that the original slice is not modified
	assert.Equal(t, []string{"foo", "bar"}, original)

	// Ensure that a new slice is returned and no overwrite occurs unlike with standard append
	original = make([]string, 0, 3)
	original = append(original, "foo", "bar")
	// Standard append has overwrite side effects
	standard := append(original, "baz")
	_ = append(original, "qux")
	assert.Equal(t, []string{"foo", "bar", "qux"}, standard)
	// Custom append has no side effects
	new1 := Append(original, "baz")
	_ = Append(original, "qux")
	assert.Equal(t, []string{"foo", "bar", "baz"}, new1)
}
