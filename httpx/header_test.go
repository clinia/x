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
		req := httptest.NewRequest("GET", "http://any.test/any", nil)
		SetCliniaHealthyHeader(req)
		h, ok := req.Header[CliniaHealthyHeaderKey]
		assert.True(t, ok)
		assert.Contains(t, h, CliniaHealthyValue)
	})
}

func TestSetCliniaUnHealthyHeader(t *testing.T) {
	t.Run("should panic on nil request", func(t *testing.T) {
		assert.Error(t, SetCliniaUnHealthyHeader(nil))
	})
	t.Run("should add the clinia healthy header as being unhealthy", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://any.test/any", nil)
		SetCliniaUnHealthyHeader(req)
		h, ok := req.Header[CliniaHealthyHeaderKey]
		assert.True(t, ok)
		assert.Contains(t, h, CliniaUnHeahlthyValue)
	})
}
