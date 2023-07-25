package funcx

// Return the first element of a function
func First[T, U any](val T, _ U) T {
	return val
}
