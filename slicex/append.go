package slicex

// Return a copy of the given slice with the given element appended. It always return a new slice.
func Append[T any, U []T](v U, e T) U {
	c := make([]T, len(v)+1)
	copy(c, v)
	c[len(v)] = e
	return c
}
