package mathx

type Number interface {
	int | int8 | int16 | int32 | int64 |
		float32 | float64 |
		uint | uint8 | uint16 | uint32 | uint64
}

// Clamp returns the value of x clamped to the range [min, max].
func Clamp[N Number](x, min, max N) N {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
