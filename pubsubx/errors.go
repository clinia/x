package pubsubx

import "errors"

type Errors []error

func (e Errors) FirstNonNil() error {
	for _, err := range e {
		if err != nil {
			return err
		}
	}

	return nil
}

func (e Errors) Join(errs ...Errors) error {
	all := e
	for _, errs := range errs {
		all = append(all, errs...)
	}

	return errors.Join(all...)
}
