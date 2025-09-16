package errorx

import (
	"testing"

	"github.com/clinia/x/pointerx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("should return if it's a clinia error", func(t *testing.T) {
		t.Run("when CliniaError", func(t *testing.T) {
			err := InternalErrorf("test")
			cErr, ok := IsCliniaError(err)
			assert.True(t, ok)
			assert.NotNil(t, cErr)
		})
		t.Run("when CliniaError ptr", func(t *testing.T) {
			err := pointerx.Ptr(InvalidArgumentErrorf("ptr test"))
			cErr, ok := IsCliniaError(err)
			assert.True(t, ok)
			assert.NotNil(t, cErr)
		})
		t.Run("not when generic error", func(t *testing.T) {
			err := errors.New("test non clinia")
			cErr, ok := IsCliniaError(err)
			assert.False(t, ok)
			assert.Nil(t, cErr)
		})
		t.Run("not when generic error message but using clinia error format", func(t *testing.T) {
			cIErr := InternalErrorf("test")
			err := errors.New(cIErr.Error())
			cErr, ok := IsCliniaError(err)
			assert.False(t, ok)
			assert.Nil(t, cErr)
		})
	})

	t.Run("should return clinia error from stack", func(t *testing.T) {
		err := NewAlreadyExistsError("test")
		serr := errors.WithStack(err)

		_, ok := IsCliniaError(serr)
		assert.True(t, ok)
	})

	t.Run("should return a clinia error without stack", func(t *testing.T) {
		err := NewAlreadyExistsError("test")

		_, ok := IsCliniaError(err)
		assert.True(t, ok)
	})

	t.Run("should return is not found from stack", func(t *testing.T) {
		err := errors.WithStack(NotFoundErrorf("test"))
		assert.True(t, IsNotFoundError(err))
	})

	t.Run("should return is not found", func(t *testing.T) {
		err := NotFoundErrorf("test")
		assert.True(t, IsNotFoundError(err))
	})

	t.Run("should return as RetryableError", func(t *testing.T) {
		err := NotFoundErrorf("test")
		retryErr := err.AsRetryableError()
		assert.Equal(t, retryErr.Unwrap(), err)
	})

	t.Run("should return as error with details", func(t *testing.T) {
		err := InternalErrorf("test")
		nfe := NotFoundError("default test")
		iae := pointerx.Ptr(InvalidArgumentErrorf("ptr test"))
		de := errors.New("test non clinia")
		wde := errors.New(nfe.Error())
		err = err.WithErrorDetails(nfe, iae, de, wde)
		assert.Len(t, err.Details, 4)
		assert.Equal(t, err.Details[0], nfe)
		assert.Equal(t, err.Details[1], *iae)
		wrapde := NewCliniaInternalErrorFromError(de)
		assert.Equal(t, err.Details[2], wrapde)
		assert.Equal(t, err.Details[3], nfe)
	})

	t.Run("should return CliniaErrors", func(t *testing.T) {
		nfe := NotFoundError("default test")
		iae := pointerx.Ptr(InvalidArgumentErrorf("ptr test"))
		de := errors.New("test non clinia")
		wde := errors.New(nfe.Error())
		errs := []error{
			nfe,
			iae,
			de,
			nil,
			wde,
		}
		wrapde := NewCliniaInternalErrorFromError(de)

		cErrs := CliniaErrorsFromErrorSlice(errs)

		assert.Len(t, cErrs, 5)
		assert.Equal(t, cErrs[0], &nfe)
		assert.Equal(t, cErrs[1], iae)
		assert.Equal(t, cErrs[2], &wrapde)
		assert.Nil(t, cErrs[3])
		assert.Equal(t, cErrs[4], &nfe)

		outErrs := cErrs.AsErrors()
		assert.Len(t, outErrs, 5)
		assert.Equal(t, outErrs[0], error(&nfe))
		assert.Equal(t, outErrs[1], error(iae))
		assert.Equal(t, outErrs[2], error(&wrapde))
		assert.Nil(t, outErrs[3])
		assert.Equal(t, outErrs[4], error(&nfe))
	})
}
