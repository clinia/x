package reflectx

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/clinia/x/loggerx"
	"github.com/clinia/x/tracex"
)

// CalculateSize calculates the size of the given value `v`.
// This function cannot panic as it will recover and log the stack trace
// using the provided logger `l` whenever a panic occurs.
//
// Parameters:
//   - l: A logger instance used for logging stack traces in case of a panic.
//   - v: The value whose size is to be calculated.
//
// Returns:
//
//	The size of the given value `v` as an integer.
func CalculateSize(ctx context.Context, l *loggerx.Logger, v interface{}) int {
	defer tracex.RecoverWithStackTrace(ctx, l, fmt.Sprintf("failed to calculate size of %T", v))
	val := reflect.ValueOf(v)
	return sizeOfValue(val, 0)
}

func sizeOfValue(val reflect.Value, depth int) int {
	if depth > 15 {
		// to avoid infinite recursion
		return 0
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return 0
		}
		return int(unsafe.Sizeof(uintptr(0))) + sizeOfValue(val.Elem(), depth+1)
	case reflect.Slice:
		size := int(unsafe.Sizeof(uintptr(0))) // size of slice header
		for i := 0; i < val.Len(); i++ {
			size += sizeOfValue(val.Index(i), depth+1)
		}
		return size
	case reflect.Map:
		size := int(unsafe.Sizeof(uintptr(0))) // size of map header
		for _, key := range val.MapKeys() {
			size += sizeOfValue(key, depth+1)
			size += sizeOfValue(val.MapIndex(key), depth+1)
		}
		return size
	case reflect.Struct:
		size := 0
		for i := 0; i < val.NumField(); i++ {
			size += sizeOfValue(val.Field(i), depth+1)
		}
		return size
	case reflect.String:
		return int(unsafe.Sizeof("")) + val.Len()
	default:
		return int(unsafe.Sizeof(val.Interface()))
	}
}
