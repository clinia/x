package stringsx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	for k, tc := range []struct {
		length int
	}{
		{
			length: 0,
		},
		{
			length: 5,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			str := Random(tc.length)
			assert.Len(t, str, tc.length)
			for i := 0; i < tc.length; i++ {
				assert.Contains(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", string(str[i]))
			}
		})
	}
}
