package featureflagx

import (
	"encoding/json"
	"slices"

	"github.com/clinia/x/errorx"
)

type FeatureFlags struct {
	fa map[FeatureFlag]FeatureFlagValue
}

func New(fs map[string]bool, ffss []FeatureFlag) (*FeatureFlags, error) {
	ffs := FeatureFlags{
		fa: map[FeatureFlag]FeatureFlagValue{},
	}
	for _, f := range ffss {
		ffs.fa[f] = boolFeatureFlagValue(false)
	}
	for f, v := range fs {
		ffs.fa[FeatureFlag(f)] = boolFeatureFlagValue(v)
	}
	err := ffs.Validate(ffss)
	if err != nil {
		return &ffs, err
	}
	return &ffs, err
}

func (ffs *FeatureFlags) IsEnabled(ff FeatureFlag) bool {
	a, ok := ffs.fa[ff]
	if !ok {
		return false
	}
	return a.IsEnabled()
}

func (ffs *FeatureFlags) GetFlags() map[FeatureFlag]FeatureFlagValue {
	return ffs.fa
}

func (ff *FeatureFlags) Validate(ffss []FeatureFlag) error {
	missingFlags := make([]FeatureFlag, 0)
	for _, f := range ffss {
		if _, ok := ff.fa[f]; !ok {
			missingFlags = append(missingFlags, f)
		}
	}
	additionalFlags := make([]FeatureFlag, 0)
	for f := range ff.fa {
		if !slices.Contains(ffss, f) {
			additionalFlags = append(additionalFlags, f)
		}
	}
	if len(missingFlags)+len(additionalFlags) > 0 {
		return errorx.InvalidArgumentErrorf("flags are missing or additional flags were provided, missing: %v, additional: %v", missingFlags, additionalFlags)
	}
	return nil
}

func (ff *FeatureFlags) MarshalJSON() ([]byte, error) {
	if len(ff.fa) == 0 {
		return []byte("{}"), nil
	}

	// Create a new map to hold the marshaled data
	marshaledData := make(map[string]bool, len(ff.fa))
	for k, v := range ff.fa {
		marshaledData[k.String()] = v.IsEnabled()
	}
	return json.Marshal(marshaledData)
}

func (ff *FeatureFlags) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var flags map[string]bool
	if err := json.Unmarshal(data, &flags); err != nil {
		return err
	}
	for k, v := range flags {
		ff.fa[FeatureFlag(k)] = boolFeatureFlagValue(v)
	}
	return nil
}
