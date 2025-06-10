package errorx

func OutputErrsMatchInputLength(errsLength, inputLength int, err error) error {
	if err != nil {
		return err
	}
	if errsLength != inputLength {
		return InternalErrorf("a different length of errors (%d) then the input length (%d) was returned", errsLength, inputLength)
	}
	return nil
}
