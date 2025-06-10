package errorx

func OutputErrsMatchInputLength(errsLength, inputLength int, errGen CliniaFunctionErrorGenerator) error {
	if errsLength != inputLength {
		return errGen("a different length of errors (%d) then the input length (%d) was returned", errsLength, inputLength)
	}
	return nil
}
