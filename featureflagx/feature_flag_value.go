package featureflagx

type FeatureFlagValue interface {
	IsEnabled() bool
}

type boolFeatureFlagValue bool

var _ FeatureFlagValue = (*boolFeatureFlagValue)(nil)

func (ffa boolFeatureFlagValue) IsEnabled() bool {
	return bool(ffa)
}
