package featureflagx

type FeatureFlagValue interface {
	IsEnabled() bool
}

type BoolFeatureFlagValue bool

var _ FeatureFlagValue = (*BoolFeatureFlagValue)(nil)

func (ffa BoolFeatureFlagValue) IsEnabled() bool {
	return bool(ffa)
}
