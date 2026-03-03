package errorx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestOutputErrsMatchInputLength(t *testing.T) {
	t.Run("should return an error on length mismatch", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{}, 2, nil)
		assert.Len(t, errs, 2)
		require.Error(t, err)
		assert.Equal(t, "[INTERNAL] a different length of errors (0) then the input length (2) was returned", err.Error())
	})

	t.Run("should return an error on length mismatch with error", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{}, 2, InvalidArgumentErrorf("test error"))
		assert.Len(t, errs, 2)
		require.Error(t, err)
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
	})
	t.Run("should return no error on length match", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{nil, errors.New("test")}, 2, nil)
		assert.NoError(t, err)
		assert.Len(t, errs, 2)
	})

	t.Run("should return error on length match with error", func(t *testing.T) {
		errs, err := OutputErrsMatchInputLength([]error{nil, errors.New("test")}, 2, InvalidArgumentErrorf("test error"))
		assert.Equal(t, "[INVALID_ARGUMENT] test error", err.Error())
		require.Error(t, err)
		assert.Len(t, errs, 2)
	})
}

func TestErrorWithAdditionalContext(t *testing.T) {
	t.Run("should be able to add additional context to an error", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithAdditionalContext("test", "ing").
			WithAdditionalContext("numerical", 10).
			WithApplicationErrorCode("TEST_CODE")

		assert.Equal(t, map[string]any{
			"test":      "ing",
			"numerical": 10,
		}, err.AdditionalContext)
		assert.Equal(t, "TEST_CODE", err.ApplicationErrorCode)
	})

	t.Run("should only take latest additional context value for a given key", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithAdditionalContext("test", "ing").
			WithAdditionalContext("test", "ed")

		assert.Equal(t, map[string]any{
			"test": "ed",
		}, err.AdditionalContext)
	})

	t.Run("should support all allowed primitive types", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("str", "hello").
			WithAdditionalContext("b", true).
			WithAdditionalContext("i", 42).
			WithAdditionalContext("i64", int64(64)).
			WithAdditionalContext("f64", float64(3.14))

		assert.Equal(t, map[string]any{
			"str": "hello",
			"b":   true,
			"i":   42,
			"i64": int64(64),
			"f64": float64(3.14),
		}, err.AdditionalContext)
	})

	t.Run("should support all allowed slice types", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("ss", []string{"a", "b"}).
			WithAdditionalContext("bs", []bool{true, false}).
			WithAdditionalContext("is", []int{1, 2}).
			WithAdditionalContext("i64s", []int64{10, 20}).
			WithAdditionalContext("f64s", []float64{1.1, 2.2})

		assert.Equal(t, map[string]any{
			"ss":   []string{"a", "b"},
			"bs":   []bool{true, false},
			"is":   []int{1, 2},
			"i64s": []int64{10, 20},
			"f64s": []float64{1.1, 2.2},
		}, err.AdditionalContext)
	})

	t.Run("should silently skip unsupported types", func(t *testing.T) {
		type custom struct{ v string }
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("valid", "kept").
			WithAdditionalContext("struct", custom{"ignored"}).
			WithAdditionalContext("ptr", &custom{"ignored"}).
			WithAdditionalContext("map", map[string]string{"k": "v"}).
			WithAdditionalContext("bytes", []byte("ignored")).
			WithAdditionalContext("int32", int32(5))

		assert.Equal(t, map[string]any{
			"valid": "kept",
		}, err.AdditionalContext)
	})

	t.Run("should overwrite an existing key with a supported type", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("key", "first").
			WithAdditionalContext("key", "second")

		assert.Equal(t, map[string]any{"key": "second"}, err.AdditionalContext)
	})

	t.Run("should not overwrite an existing key when new value type is unsupported", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("key", "original").
			WithAdditionalContext("key", struct{}{})

		assert.Equal(t, map[string]any{"key": "original"}, err.AdditionalContext)
	})

	t.Run("should initialise the map on nil AdditionalContext", func(t *testing.T) {
		err := InvalidArgumentErrorf("test")
		assert.Nil(t, err.AdditionalContext)

		err = err.WithAdditionalContext("k", "v")
		assert.NotNil(t, err.AdditionalContext)
		assert.Equal(t, map[string]any{"k": "v"}, err.AdditionalContext)
	})

	t.Run("should not mutate original error (value receiver copy semantics)", func(t *testing.T) {
		original := InvalidArgumentErrorf("test")
		_ = original.WithAdditionalContext("k", "v")

		assert.Nil(t, original.AdditionalContext)
	})

	t.Run("should only take latest application error code", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error").
			WithApplicationErrorCode("TEST_CODE").
			WithApplicationErrorCode("TEST_CODE_2")

		assert.Equal(t, "TEST_CODE_2", err.ApplicationErrorCode)
	})

	t.Run("should be able to attach additional details on an error pointer", func(t *testing.T) {
		err := InvalidArgumentErrorf("test error")
		errPtr := &err
		err = errPtr.WithAdditionalContext("test", "ing").
			WithAdditionalContext("numerical", 10).
			WithApplicationErrorCode("TEST_CODE")

		assert.Equal(t, map[string]any{
			"test":      "ing",
			"numerical": 10,
		}, err.AdditionalContext)
		assert.Equal(t, "TEST_CODE", err.ApplicationErrorCode)
	})
}

func TestWithAdditionalContextMap(t *testing.T) {
	t.Run("should apply all entries from a single map", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextMap(map[string]any{
				"a": "1",
				"b": true,
			})

		assert.Equal(t, map[string]any{"a": "1", "b": true}, err.AdditionalContext)
	})

	t.Run("should skip unsupported types within the map", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextMap(map[string]any{
				"valid":   "yes",
				"invalid": struct{}{},
				"map":     map[string]string{"k": "v"},
			})

		assert.Equal(t, map[string]any{"valid": "yes"}, err.AdditionalContext)
	})

	t.Run("later map values should override earlier ones for the same key", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextMap(
				map[string]any{"k": "first"},
				map[string]any{"k": "second"},
			)

		assert.Equal(t, map[string]any{"k": "second"}, err.AdditionalContext)
	})

	t.Run("should not override existing key when later map value type is unsupported", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextMap(
				map[string]any{"k": "original"},
				map[string]any{"k": struct{}{}},
			)

		assert.Equal(t, map[string]any{"k": "original"}, err.AdditionalContext)
	})

	t.Run("should accept no maps (no-op)", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").WithAdditionalContextMap()
		assert.Nil(t, err.AdditionalContext)
	})

	t.Run("should accept an empty map (no-op)", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").WithAdditionalContextMap(map[string]any{})
		assert.Nil(t, err.AdditionalContext)
	})

	t.Run("should merge with pre-existing context", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("existing", "value").
			WithAdditionalContextMap(map[string]any{"new": "entry"})

		assert.Equal(t, map[string]any{
			"existing": "value",
			"new":      "entry",
		}, err.AdditionalContext)
	})
}

func TestWithAdditionalContextAttributes(t *testing.T) {
	t.Run("should support all OTel attribute types", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextAttributes(
				attribute.String("str", "hello"),
				attribute.Bool("b", true),
				attribute.Int64("i64", 99),
				attribute.Float64("f64", 1.5),
				attribute.StringSlice("ss", []string{"x", "y"}),
				attribute.BoolSlice("bs", []bool{false, true}),
				attribute.Int64Slice("i64s", []int64{1, 2}),
				attribute.Float64Slice("f64s", []float64{0.1, 0.2}),
			)

		assert.Equal(t, map[string]any{
			"str":  "hello",
			"b":    true,
			"i64":  int64(99),
			"f64":  float64(1.5),
			"ss":   []string{"x", "y"},
			"bs":   []bool{false, true},
			"i64s": []int64{1, 2},
			"f64s": []float64{0.1, 0.2},
		}, err.AdditionalContext)
	})

	t.Run("should overwrite an existing key", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContextAttributes(
				attribute.String("k", "first"),
				attribute.String("k", "second"),
			)

		assert.Equal(t, map[string]any{"k": "second"}, err.AdditionalContext)
	})

	t.Run("should merge with pre-existing context", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").
			WithAdditionalContext("pre", "existing").
			WithAdditionalContextAttributes(attribute.String("attr", "value"))

		assert.Equal(t, map[string]any{
			"pre":  "existing",
			"attr": "value",
		}, err.AdditionalContext)
	})

	t.Run("should accept no attributes (no-op)", func(t *testing.T) {
		err := InvalidArgumentErrorf("test").WithAdditionalContextAttributes()
		assert.Nil(t, err.AdditionalContext)
	})

	t.Run("should not mutate the original error", func(t *testing.T) {
		original := InvalidArgumentErrorf("test")
		_ = original.WithAdditionalContextAttributes(attribute.String("k", "v"))

		assert.Nil(t, original.AdditionalContext)
	})
}
