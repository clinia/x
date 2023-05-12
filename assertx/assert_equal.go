package assertx

import (
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func Equal(t assert.TestingT, expected interface{}, actual interface{}, opts ...cmp.Option) (ok bool) {
	if !cmp.Equal(expected, actual, opts...) {
		t.Errorf("Not equal: \n%s", cmp.Diff(expected, actual, opts...))
		return false
	}

	return true
}
