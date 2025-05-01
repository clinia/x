package featureflagx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolFeatureFlagValue(t *testing.T) {
	t.Run("should return its own bool value", func(t *testing.T) {
		ba1 := boolFeatureFlagValue(true)
		assert.True(t, ba1.IsEnabled())
		ba2 := boolFeatureFlagValue(false)
		assert.False(t, ba2.IsEnabled())
	})
}
