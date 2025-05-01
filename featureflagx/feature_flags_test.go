package featureflagx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlags(t *testing.T) {
	Flags := []FeatureFlag{
		FeatureFlag("feature1"),
		FeatureFlag("feature2"),
		FeatureFlag("feature3"),
	}
	t.Run("should return valid if all flags are included", func(t *testing.T) {
		ffm := map[string]bool{}
		for _, f := range Flags {
			ffm[f.String()] = true
		}
		ff, err := New(ffm, Flags)
		assert.NoError(t, err)
		assert.NotNil(t, ff)
		assert.Equal(t, len(ff.fa), len(Flags))
		assert.Equal(t, map[FeatureFlag]FeatureFlagValue{
			Flags[0]: boolFeatureFlagValue(true),
			Flags[1]: boolFeatureFlagValue(true),
			Flags[2]: boolFeatureFlagValue(true),
		}, ff.fa)
	})

	t.Run("should return invalid if some flags are missing", func(t *testing.T) {

		ffm := map[string]bool{}
		for _, f := range Flags {
			ffm[f.String()] = true
		}
		ff, err := New(ffm, Flags)
		assert.NoError(t, err)
		delete(ff.fa, Flags[0])
		assert.Error(t, ff.Validate(Flags))
	})

	t.Run("should return invalid if some flags are additional", func(t *testing.T) {
		ffm := map[string]bool{}
		for _, f := range Flags {
			ffm[f.String()] = true
		}
		ff, err := New(ffm, Flags)
		assert.NoError(t, err)
		ff.fa[FeatureFlag("feature4")] = boolFeatureFlagValue(true)
		assert.Error(t, ff.Validate(Flags))
	})

	t.Run("should disable a flag when set to false", func(t *testing.T) {
		ffm := map[string]bool{
			"feature1": true,
			"feature2": false,
		}
		ff, err := New(ffm, Flags)
		assert.NoError(t, err)
		assert.NotNil(t, ff)
		assert.False(t, ff.IsEnabled(FeatureFlag("feature2")))
		assert.True(t, ff.IsEnabled(FeatureFlag("feature1")))
	})

	t.Run("should correctly marshal feature flags to JSON", func(t *testing.T) {
		ffm := map[string]bool{
			"feature1": true,
			"feature2": false,
			"feature3": true,
		}
		ff, err := New(ffm, Flags)
		assert.NoError(t, err)
		assert.NotNil(t, ff)

		expectedJSON := `{"feature1":true,"feature2":false,"feature3":true}`
		marshaled, err := ff.MarshalJSON()
		assert.NoError(t, err)
		assert.JSONEq(t, expectedJSON, string(marshaled))
	})

	t.Run("should correctly unmarshal JSON into feature flags", func(t *testing.T) {
		jsonData := `{"feature1":true,"feature2":false,"feature3":true}`
		ff := &FeatureFlags{
			fa: map[FeatureFlag]FeatureFlagValue{},
		}
		err := ff.UnmarshalJSON([]byte(jsonData))
		assert.NoError(t, err)
		assert.False(t, ff.IsEnabled(FeatureFlag("feature2")))
		assert.True(t, ff.IsEnabled(FeatureFlag("feature1")))
		assert.True(t, ff.IsEnabled(FeatureFlag("feature3")))
		assert.Equal(t, len(ff.GetFlags()), 3)
	})
}
