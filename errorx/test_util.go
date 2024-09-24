package errorx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// PanicsWithCliniaError asserts that the provided function panics with the expected CliniaError.
// This is functionally equivalent to require.PanicsWithValue, but with a CliniaError (since our error type contains slices which are not comparable with the `==` operator).
func PanicsWithCliniaError(t *testing.T, expected CliniaError, f func()) {
	t.Helper()
	fail := func() { t.Fatalf("expected panic with error, but function did not panic") }
	defer func() {
		if r := recover(); r != nil {
			require.Equal(t, expected, r)
		} else {
			fail()
		}
	}()

	f()
}
