package utilx

/*
This if for functions that returns the patter `func(...) (out T, err error)` and is called in a function
that expect this to not return any error and is also not returning error. It fallback to panic if an error is found.

	obj, err := MyFunc()
	require.NoError(t, err)
	s := MyStruct{
		objField: obj
	}

becomes

	s := MyStruct{
		objField: utilx.Must(MyFunc())
	}
*/
func Must[T any](item T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return item
}
