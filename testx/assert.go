package testx

func AssertAndPanicOnError[T any](item T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return item
}
