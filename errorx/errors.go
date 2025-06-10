package errorx

func OutputErrsMatchInputLength(errs []error, inputLength int, err error) ([]error, error) {
	if err != nil {
		return make([]error, inputLength), err
	}
	if len(errs) != inputLength {
		return make([]error, inputLength), InternalErrorf("a different length of errors (%d) then the input length (%d) was returned", len(errs), inputLength)
	}
	return errs, nil
}
