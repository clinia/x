package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	t.Parallel()

	// Test floats
	{
		for _, test := range []struct {
			v, low, high, expected float32
		}{
			{
				v:        0.5,
				low:      0,
				high:     1,
				expected: 0.5,
			},
			{
				v:        -0.5,
				low:      0,
				high:     1,
				expected: 0,
			},
			{
				v:        1.5,
				low:      0,
				high:     1,
				expected: 1,
			},

			{
				v:        1.5,
				low:      1,
				high:     2,
				expected: 1.5,
			},
			{
				v:        0.5,
				low:      1,
				high:     2,
				expected: 1,
			},
			{
				v:        2.5,
				low:      1,
				high:     2,
				expected: 2,
			},
			{
				v:        0.999999,
				low:      1,
				high:     2,
				expected: 1,
			},
			{
				v:        2.000001,
				low:      1,
				high:     2,
				expected: 2,
			},
			{
				v:        1.000001,
				low:      1,
				high:     2,
				expected: 1.000001,
			},
		} {
			assert.Equal(t, test.expected, Clamp(test.v, test.low, test.high))
		}
	}

	// Test ints
	{
		for _, test := range []struct {
			v, low, high, expected int
		}{
			{
				v:        5,
				low:      0,
				high:     10,
				expected: 5,
			},
			{
				v:        -5,
				low:      0,
				high:     10,
				expected: 0,
			},
			{
				v:        15,
				low:      0,
				high:     10,
				expected: 10,
			},
			{
				v:        10,
				low:      0,
				high:     10,
				expected: 10,
			},
			{
				v:        0,
				low:      0,
				high:     10,
				expected: 0,
			},
		} {
			assert.Equal(t, test.expected, Clamp(test.v, test.low, test.high))
		}
	}
}
