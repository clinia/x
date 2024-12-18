package mathx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesToMB(t *testing.T) {
	tests := []struct {
		input    int
		expected float64
	}{
		{1048576, 1},   // 1 MB
		{2097152, 2},   // 2 MB
		{5242880, 5},   // 5 MB
		{0, 0},         // 0 bytes
		{10485760, 10}, // 10 MB
	}

	for _, test := range tests {
		result := BytesToMB(test.input)
		require.Equal(t, test.expected, result, "BytesToMB(%d) = %f; expected %f", test.input, result, test.expected)
	}
}
