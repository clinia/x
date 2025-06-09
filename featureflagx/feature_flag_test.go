package featureflagx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlag(t *testing.T) {
	t.Run("should return itself when retrieving feature flag name", func(t *testing.T) {
		f := FeatureFlag("ff")
		assert.Equal(t, f, f.String())
	})
}
