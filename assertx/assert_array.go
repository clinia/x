// The code has been taken from "github.com/stretchr/testify/assert" but with a slight
// modification to use go-cmp instead of reflect in order to support additional
// options like IgnoreFields.
package assertx

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

type tHelper interface {
	Helper()
}

var spewConfig = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	DisableMethods:          true,
	MaxDepth:                10,
}

// ElementsMatch asserts that the specified listA(array, slice...) is equal to specified
// listB(array, slice...) ignoring the order of the elements. If there are duplicate elements,
// the number of appearances of each of them in both lists should match.
//
// assert.ElementsMatch(t, [1, 3, 2, 3], [1, 3, 3, 2])
func ElementsMatch(t assert.TestingT, listA, listB interface{}, opts ...cmp.Option) (ok bool) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if isEmpty(listA) && isEmpty(listB) {
		return true
	}

	if !isList(t, listA) || !isList(t, listB) {
		return false
	}

	extraA, extraB := diffLists(listA, listB, opts...)

	if len(extraA) == 0 && len(extraB) == 0 {
		return true
	}

	return assert.Fail(t, formatListDiff(listA, listB, extraA, extraB))
}

// isEmpty gets whether the specified object is considered empty or not.
func isEmpty(object interface{}) bool {
	// get nil case out of the way
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
	// for all other types, compare against the zero value
	// array types are empty when they match their zero-initialized state
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

// isList checks that the provided value is array or slice.
func isList(t assert.TestingT, list interface{}) (ok bool) {
	kind := reflect.TypeOf(list).Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		return assert.Fail(t, fmt.Sprintf("%q has an unsupported type %s, expecting array or slice", list, kind))
	}
	return true
}

// diffLists diffs two arrays/slices and returns slices of elements that are only in A and only in B.
// If some element is present multiple times, each instance is counted separately (e.g. if something is 2x in A and
// 5x in B, it will be 0x in extraA and 3x in extraB). The order of items in both lists is ignored.
func diffLists(listA, listB interface{}, opts ...cmp.Option) (extraA, extraB []interface{}) {
	aValue := reflect.ValueOf(listA)
	bValue := reflect.ValueOf(listB)

	aLen := aValue.Len()
	bLen := bValue.Len()

	// Mark indexes in bValue that we already used
	visited := make([]bool, bLen)
	for i := 0; i < aLen; i++ {
		element := aValue.Index(i).Interface()
		found := false
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}

			if cmp.Equal(bValue.Index(j).Interface(), element, opts...) {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			extraA = append(extraA, element)
		}
	}

	for j := 0; j < bLen; j++ {
		if visited[j] {
			continue
		}
		extraB = append(extraB, bValue.Index(j).Interface())
	}

	return
}

func formatListDiff(listA, listB interface{}, extraA, extraB []interface{}) string {
	var msg bytes.Buffer

	msg.WriteString("elements differ")
	if len(extraA) > 0 {
		msg.WriteString("\n\nextra elements in list A:\n")
		msg.WriteString(spewConfig.Sdump(extraA))
	}
	if len(extraB) > 0 {
		msg.WriteString("\n\nextra elements in list B:\n")
		msg.WriteString(spewConfig.Sdump(extraB))
	}
	msg.WriteString("\n\nlistA:\n")
	msg.WriteString(spewConfig.Sdump(listA))
	msg.WriteString("\n\nlistB:\n")
	msg.WriteString(spewConfig.Sdump(listB))

	return msg.String()
}
