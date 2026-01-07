package reflectx

import (
	"testing"

	loggerxtest "github.com/clinia/x/loggerx/test"
	"github.com/stretchr/testify/require"
)

func TestSizeOf(t *testing.T) {
	l := loggerxtest.NewTestLogger(t)

	type myOtherStruct struct {
		A    int
		B    string
		Data []byte
	}

	type myStruct struct {
		A      int
		B      string
		Data1  []byte
		Nested *myOtherStruct
		Others []*myOtherStruct
	}

	t.Run("should give the approximate size of a struct", func(t *testing.T) {
		data := make([]byte, 1*1024*1024) // 1MB

		mockedStruct := myStruct{
			A:     1,
			B:     "hello",
			Data1: data,
			Nested: &myOtherStruct{
				A:    2,
				B:    "world",
				Data: data,
			},
			Others: []*myOtherStruct{
				{
					A:    3,
					B:    "foo",
					Data: data,
				},
				{
					A:    4,
					B:    "bar",
					Data: data,
				},
			},
		}

		size := CalculateSize(t.Context(), l, mockedStruct)
		t.Logf("Size of myStruct: %d bytes", size)
		require.GreaterOrEqual(t, size, 4*1024*1024) // 4MB
	})

	t.Run("should avoid infinite recursion", func(t *testing.T) {
		type infiniteRecursionStruct struct {
			Next *infiniteRecursionStruct
		}

		mockedStruct := &infiniteRecursionStruct{}
		mockedStruct.Next = mockedStruct

		size := CalculateSize(t.Context(), l, mockedStruct)
		require.GreaterOrEqual(t, size, 0)
	})
}
