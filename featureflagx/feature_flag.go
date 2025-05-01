package featureflagx

//	type FeatureFlag interface {
//		FeatureFlagName() featureFlag
//		String() string
//	}
type FeatureFlag string

// func NewFeatureFlag(ff string) FeatureFlag {
// 	return featureFlag(ff)
// }

// func (ff FeatureFlag) FeatureFlagName() FeatureFlag {
// 	return ff
// }

func (ff FeatureFlag) String() string {
	return string(ff)
}
