package errorx

/*
This is to reduce the boilerplate check, it checks the length of the error slice with the expected inputLength.
Also takes in an error since the usual pattern for batch function is to look like `func(ctx context.Context, inputs []any) ([]any, []error, error)`
This takes the priority of if :
- There is an error, return the error back
- There is no error but the length of the errs slice mismatch the inputLength, return a preformatted error
- Return nil if no problem are found

	outputs, errs, err := funcThatProcess(ctx, inputs)
	if err != nil {
	return nil, make(error[], len(inputs)), err
	}
	if len(errs) != len(inputs) {
	return nil, make(error[], len(inputs)), errorx.InternalErrorf("inputs mismatch")
	}
	return outputs, errs, nil

and we can replace this with:

	outputs, errs, err := funcThatProcess(ctx, inputs)
	if errs, err := OutputErrsMatchInputLength(errs, len(inputs), err); err != nil {
	   return nil, errs, err
	}
	return outputs, errs, nil
*/
func OutputErrsMatchInputLength(errs []error, inputLength int, err error) ([]error, error) {
	if err != nil {
		return make([]error, inputLength), err
	}
	if len(errs) != inputLength {
		return make([]error, inputLength), InternalErrorf("a different length of errors (%d) then the input length (%d) was returned", len(errs), inputLength)
	}
	return errs, nil
}
