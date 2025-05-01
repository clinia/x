package featureflagx

import (
	"context"
)

type featureFlagKeyType uint8

var featureFlagKey = featureFlagKeyType(1)

func NewContext(ctx context.Context, ff *FeatureFlags) context.Context {
	return context.WithValue(ctx, featureFlagKey, ff)
}

func FromContext(ctx context.Context) (*FeatureFlags, bool) {
	val := ctx.Value(featureFlagKey)
	ff, ok := val.(*FeatureFlags)
	if !ok {
		return nil, false
	}
	return ff, true
}
