package errorx

func OutputErrsMatchInputLength(errsLength, inputLength int) error {
	if errsLength != inputLength {
		return InternalErrorf("a different length of errors (%d) then the input length (%d) was returned", errsLength, inputLength)
	}
	return nil
}
