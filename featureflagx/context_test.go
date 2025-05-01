package featureflagx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagContext(t *testing.T) {
	Flags := []FeatureFlag{
		FeatureFlag("feature1"),
		FeatureFlag("feature2"),
		FeatureFlag("feature3"),
	}
	t.Run("should add the feature flag to the context and be able to retrieve it", func(t *testing.T) {
		ctx := context.Background()
		ff, err := New(
			map[string]bool{
				string("feature1"): false,
				string("feature2"): true,
				string("feature3"): false,
			},
			Flags,
		)
		require.NoError(t, err)
		uctx := NewContext(ctx, ff)
		ffctx, ok := FromContext(uctx)
		assert.NotEqual(t, ctx, uctx)
		assert.True(t, ok)
		assert.Equal(t, ff, ffctx)
	})

	t.Run("should return false on empty value in feature flag", func(t *testing.T) {
		ctx := context.Background()
		_, ok := FromContext(ctx)
		assert.False(t, ok)
	})
}
