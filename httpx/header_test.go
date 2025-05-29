package httpx

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCliniaHealthyHeader(t *testing.T) {
	t.Run("should panic on nil request", func(t *testing.T) {
		assert.Error(t, SetCliniaHealthyHeader(nil))
	})
	t.Run("should add the clinia healthy header as being healthy", func(t *testing.T) {
		w := httptest.NewRecorder()
		_ = SetCliniaHealthyHeader(w)
		h := w.Header().Get(CliniaHealthyHeaderKey)
		assert.Contains(t, h, CliniaHealthyValue)
	})
}

func TestSetCliniaUnHealthyHeader(t *testing.T) {
	t.Run("should panic on nil request", func(t *testing.T) {
		assert.Error(t, SetCliniaUnHealthyHeader(nil))
	})
	t.Run("should add the clinia healthy header as being unhealthy", func(t *testing.T) {
		w := httptest.NewRecorder()
		_ = SetCliniaUnHealthyHeader(w)
		h := w.Header().Get(CliniaHealthyHeaderKey)
		assert.Contains(t, h, CliniaUnHealthyValue)
	})
}
