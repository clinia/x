package slicex

// Return a copy of the given slice with the given element appended. It always return a new slice.
func Append[T any](v []T, e T) []T {
	c := make([]T, len(v), len(v)+1)
	copy(c, v)
	return append(c, e)
}
